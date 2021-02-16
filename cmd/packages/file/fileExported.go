package file

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"

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

// Peer struct containg ip and port of the client
type Peer struct {
	IP   net.IP
	Port uint16
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
	err = t.requestPeers(Pid, 7000)
	if err != nil {
		return err
	}
	return nil
}

//DecodePeer takes the peers string and converts them into the an array of Peers struct
func DecodePeer(bin []byte) ([]Peer, error) {
	const size = 6
	num := len(bin) / size
	if len(bin)%size != 0 {
		err := fmt.Errorf("Peers string has been corrupted")
		return nil, err
	}
	peers := make([]Peer, num)

	for i := 0; i < num; i++ {
		offset := i * size
		peers[i].IP = net.IP(bin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(bin[offset+4 : offset+6]))
	}
	fmt.Println("helo")
	fmt.Println(peers)
	fmt.Println("he")

	return peers, nil
}

//String is used to convert the peer struct to a valid ip address
func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
