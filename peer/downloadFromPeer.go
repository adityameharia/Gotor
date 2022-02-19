package peer

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	connection "github.com/adityameharia/gotor/connection"
	message "github.com/adityameharia/gotor/message"
	"log"
	"runtime"
	"time"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

type work struct {
	index  int
	hash   [20]byte
	length int
}

type result struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *connection.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

type pieceResult struct {
	index int
	buf   []byte
}

//Download make a worker queues,launches go routines ,co-ordinates with the result
//basically this func is the heart of the package which does it all
func (t *Torrent) Download() ([]byte, error) {

	log.Println("Starting download for", t.Name)

	workerQueue := make(chan *work, len(t.PieceHashes))
	workerResults := make(chan *result)

	for index, hash := range t.PieceHashes {
		length := t.pieceSize(index)
		workerQueue <- &work{index, hash, length}
	}

	// Start workers
	for _, peer := range t.Peers {
		go t.startDownload(peer, workerQueue, workerResults)
	}

	// Collect results into a buffer until full
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		fmt.Println("a piece done")
		res := <-workerResults
		begin, end := t.calculateBounds(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		fmt.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	close(workerQueue)

	return buf, nil

}

func (t *Torrent) calculateBounds(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) pieceSize(index int) int {
	b := index * t.PieceLength
	e := b + t.PieceLength
	if e > t.Length {
		e = t.Length
	}
	return e - b
}

func (t *Torrent) startDownload(peer Peer, workQueue chan *work, results chan *result) {
	c, err := connection.New(peer.String(), t.PeerID, t.InfoHash)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.Conn.Close()

	log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQueue {
		if !c.Bitfield.CheckPiece(pw.index) {
			workQueue <- pw // Put piece back on the queue
			continue
		}

		fmt.Println("hi")

		// Download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}
		fmt.Println("bye")
		err = checkIntegrity(pw, buf)
		fmt.Println("ffs")
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw // Put piece back on the queue
			continue
		}

		c.SendHave(pw.index)
		results <- &result{pw.index, buf}
	}
}

func attemptDownloadPiece(c *connection.Client, pw *work) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pw.length {
		// If unchoked, send requests until we have enough unfulfilled requests
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				// Last block might be shorter than the typical block
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.parseMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}

func (p *pieceProgress) parseMessage() error {
	msg, err := p.client.ReadMessageFromPeer()
	if err != nil {
		return err
	}

	if msg == nil { // keep-alive
		return nil
	}

	switch msg.ID {
	case message.Unchoke:
		p.client.Choked = false
	case message.Choke:
		p.client.Choked = true
	case message.Have:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		p.client.Bitfield.PutPiece(index)
	case message.Piece:
		n, err := message.ParsePiece(p.index, p.buf, msg)
		if err != nil {
			return err
		}
		p.downloaded += n
		p.backlog--
	}
	return nil
}

func checkIntegrity(pw *work, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}
