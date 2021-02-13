package file

import (
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type benInfo struct {
	Announce    string `bencode:"announce"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `benocode:"length"`
	Name        string `bencode:"name`
}

func Open(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()
	b := benInfo{}
	err = bencode.Unmarshal(f, &b)
	if err != nil {
		return err
	}
	fmt.Println(b)
	return nil
}
