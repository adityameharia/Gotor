package peer

import (
	"fmt"
	connection "gotor/cmd/packages/connection"
	"log"
)

type work struct {
	index  int
	hash   [20]byte
	length int
}

type result struct {
	index int
	buf   []byte
}

//Download make a worker queues,launches go routines ,co-ordinates with the result
//basically this func is the heart of the package which does it all
func (t *Torrent) Download() error {
	workerQueue := make(chan *work, len(t.PieceHashes))
	//WorkerResults := make(chan *result)

	for index, hash := range t.PieceHashes {
		length := t.pieceSize(index)
		workerQueue <- &work{index, hash, length}
	}
	<-workerQueue
	fmt.Println(<-workerQueue)
	t.startDownload(t.Peers[1], workerQueue)
	return nil
}

func (t *Torrent) pieceSize(index int) int {
	b := index * t.PieceLength
	e := b + t.PieceLength
	if e > t.Length {
		e = t.Length
	}
	return e - b
}

func (t *Torrent) startDownload(peer Peer, workQueue chan *work) {
	c, err := connection.New(peer.String(), t.PeerID, t.InfoHash)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.Conn.Close()

	log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()
}
