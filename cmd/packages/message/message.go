//Package message has been literally copied code from veggiefender/torrent-client
package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

// Message stores ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
}

const (
	// Choke chokes the receiver
	Choke messageID = 0
	// Unchoke unchokes the receiver
	Unchoke messageID = 1
	// Interested expresses interest in receiving data
	Interested messageID = 2
	// NotInterested expresses disinterest in receiving data
	NotInterested messageID = 3
	// Have alerts the receiver that the sender has downloaded a piece
	Have messageID = 4
	// Bitfield encodes which pieces that the sender has downloaded
	Bitfield messageID = 5
	// Request requests a block of data from the receiver
	Request messageID = 6
	// Piece delivers a block of data to fulfill a request
	Piece messageID = 7
	// Cancel cancels a request
	Cancel messageID = 8
)

//Read reads a message from stream.
//The first 4 bytes of the stream gives the length of the message and hence we get the length then read that many bytes from string.
//Return nil on keep alive msg,i.e to not close the connection
func Read(r io.Reader) (*Message, error) {
	buf := make([]byte, 4)

	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	l := binary.BigEndian.Uint32(buf)
	if l == 0 {
		return nil, nil
	}

	msg := make([]byte, l)
	_, err = io.ReadFull(r, msg)
	if err != nil {
		return nil, err
	}

	m := Message{
		ID:      messageID(msg[0]),
		Payload: msg[1:],
	}

	return &m, nil
}

func (m *Message) String() string {
	if m == nil {
		return m.name()
	}
	return fmt.Sprintf("%s [%d]", m.name(), len(m.Payload))
}

func (m *Message) name() string {
	if m == nil {
		return "KeepAlive"
	}
	switch m.ID {
	case Choke:
		return "Choke"
	case Unchoke:
		return "Unchoke"
	case Interested:
		return "Interested"
	case NotInterested:
		return "NotInterested"
	case Have:
		return "Have"
	case Bitfield:
		return "Bitfield"
	case Request:
		return "Request"
	case Piece:
		return "Piece"
	case Cancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown#%d", m.ID)
	}
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	length := uint32(len(m.Payload) + 1) // +1 for id
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

// FormatRequest creates a REQUEST message
func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: Request, Payload: payload}
}

// ParseHave parses a HAVE message
func ParseHave(msg *Message) (int, error) {
	if msg.ID != Have {
		return 0, fmt.Errorf("Expected HAVE (ID %d), got ID %d", Have, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("Expected payload length 4, got length %d", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

// ParsePiece parses a PIECE message and copies its payload into a buffer
func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != Piece {
		return 0, fmt.Errorf("Expected PIECE (ID %d), got ID %d", Piece, msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("Payload too short. %d < 8", len(msg.Payload))
	}
	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("Expected index %d, got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin offset too high. %d >= %d", begin, len(buf))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("Data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
	}
	copy(buf[begin:], data)
	return len(data), nil
}

// FormatHave creates a HAVE message
func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: Have, Payload: payload}
}
