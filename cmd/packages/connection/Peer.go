package connection

import (
	message "gotor/cmd/packages/message"
)

// SendUnchoke sends an Unchoke message to the peer
func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.Unchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendInterested sends an Interested message to the peer
func (c *Client) SendInterested() error {
	msg := message.Message{ID: message.Interested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendRequest sends a Request message to the peer
func (c *Client) SendRequest(index, begin, length int) error {
	req := message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

//ReadMessageFromPeer read the peers response after handshake,i.e interprets have,choked,unchoked,and Piece messages
func (c *Client) ReadMessageFromPeer() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

// SendHave sends a Have message to the peer
func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
