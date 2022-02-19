package file

import (
	"crypto/rand"
	"fmt"
	peer "github.com/adityameharia/gotor/peer"
	"os"

	"github.com/jackpal/bencode-go"
)

//TorrentFile struct of the torrent file
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

//Tracker has the peers string and the time interval after which to send another request
type Tracker struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

//Open is used to open the file,unmarshall the contents of the file and convert it to the form of a torrentFile
func Open(path string) (TorrentFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}

	defer f.Close()
	b := bencodeTorrent{}
	err = bencode.Unmarshal(f, &b)
	if err != nil {
		return TorrentFile{}, err
	}
	return b.toTorrentFile()
}

//DownloadFile is used generate a random id for us to be identified with and get a list of all the peer swith their ips and ports
func (t *TorrentFile) DownloadFile(path string) error {
	Pid := make([]byte, 20)
	_, err := rand.Read(Pid)
	if err != nil {
		return err
	}
	peers, err := t.requestPeers(Pid, 7000)
	if err != nil {
		return err
	}

	fmt.Println(peers)

	torrent := peer.Torrent{
		Peers:       peers,
		PeerID:      Pid,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(outFile)

	buf, err := torrent.Download()

	if err != nil {
		return err
	}

	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}
