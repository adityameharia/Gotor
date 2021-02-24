package connection

import (
	"fmt"
	message "gotor/cmd/packages/message"
	"io"
	"net"
	"time"
)

type peer struct {
	IP   net.IP
	Port uint16
}

type handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   []byte
}

//Bitfield is byte array which stores the index of the parts available with a particular client
type Bitfield []byte

// A Client is a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield Bitfield
	peer     string
	infoHash [20]byte
	peerID   []byte
}

// CheckPiece tells if a bitfield has a particular index set
func (bf Bitfield) CheckPiece(index int) bool {
	number := index / 8
	off := index % 8
	if number < 0 || number >= len(bf) {
		return false
	}
	return bf[number]>>(7-off)&1 != 0
}

// PutPiece sets a bit in the bitfield
func (bf Bitfield) PutPiece(index int) {
	number := index / 8
	off := index % 8

	// silently discard invalid bounded index
	if number < 0 || number >= len(bf) {
		return
	}
	bf[number] |= 1 << (7 - off)
}

// New connects with a peer, completes a handshake, and receives a handshake
// returns an err if any of those fail.
func New(peer string, pid []byte, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer, 3*time.Second)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = peerHandshake(conn, infoHash, pid)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bitf, err := manipulateBitfield(conn)

	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bitf,
		peer:     peer,
		infoHash: infoHash,
		peerID:   pid,
	}, nil
}

func peerHandshake(conn net.Conn, infohash [20]byte, Pid []byte) error {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	req := handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infohash,
		PeerID:   Pid,
	}
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return err
	}

	res, err := req.Read(conn)
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}

func (h *handshake) Serialize() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}

func (h *handshake) Read(r io.Reader) (*handshake, error) {
	length := make([]byte, 1)
	_, err := io.ReadFull(r, length)
	if err != nil {
		return nil, err
	}

	resLen := int(length[0])

	if resLen == 0 {
		err := fmt.Errorf("resLen cannot be 0")
		return nil, err
	}

	buffer := make([]byte, 48+resLen)
	_, err = io.ReadFull(r, buffer)
	if err != nil {
		return nil, err
	}

	var infoHash [20]byte
	var peerID []byte

	copy(infoHash[:], buffer[resLen+8:resLen+8+20])
	copy(peerID[:], buffer[resLen+8+20:])

	return &handshake{
		Pstr:     string(buffer[0:resLen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil

}

func manipulateBitfield(c net.Conn) (Bitfield, error) {
	c.SetDeadline(time.Now().Add(5 * time.Second))

	//we disbale the deadline if we get a valid resp
	defer c.SetDeadline(time.Time{})

	msg, err := message.Read(c)

	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("Expected bitfield but got %s", msg)
		return nil, err
	}

	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}
