package peer

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

//MaxSize is the maximmum size we get request for from a peer in one request
const MaxSize = 16384

//MaxInPipeLine is the maximum number of request that can be kept in queue waiting for them to be picked up by a worker
const MaxInPipeLine = 5

// Torrent holds data required to download a torrent from a list of peers
type Torrent struct {
	Peers       []Peer
	PeerID      []byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

// Peer struct containg ip and port of the client
type Peer struct {
	IP   net.IP
	Port uint16
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

	//fmt.Println(peers)

	return peers, nil
}

//String is used to convert the peer struct to a valid ip address
func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
