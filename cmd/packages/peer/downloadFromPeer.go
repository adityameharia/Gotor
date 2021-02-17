package peer

import "fmt"

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
