package file

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	peer "github.com/adityameharia/gotor/peer"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

//request peers takes the announce url in the torrent file and adds a few url encoded parameters to it.
//It then decodes the peers binary blob recieved as a response to an array of Peer structs
func (t *TorrentFile) requestPeers(Pid []byte, port uint16) ([]peer.Peer, error) {
	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(Pid[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	Requrl := t.Announce + "?" + params.Encode()

	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(Requrl)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	tracker := Tracker{}
	err = bencode.Unmarshal(resp.Body, &tracker)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// base, err := url.Parse(t.Announce)
	// if err != nil {
	// 	return nil, err
	// }

	// CONNECT := base.Host

	// s, err := net.ResolveUDPAddr("udp4", CONNECT)
	// c, err := net.DialUDP("udp4", nil, s)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, err
	// }
	// fmt.Println("hi")
	// var buf []byte
	// fmt.Println("bye")
	// n, _, err := c.ReadFromUDP(buf)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, err
	// }
	// fmt.Println("hel")
	// fmt.Printf("Reply: %s\n", string(buf[0:n]))
	return peer.DecodePeer([]byte(tracker.Peers))
	// fmt.Println(c.RemoteAddr().String())
	// return nil, nil
}

//toTorrentFile converts the bencode torrent to a torrentFile struct
func (b *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, b.Info)
	if err != nil {
		return TorrentFile{}, err
	}

	h := sha1.Sum(buf.Bytes())
	leng := 20
	piece := []byte(b.Info.Pieces)
	if len(piece)%leng != 0 {
		err := fmt.Errorf("Torrent File is corrupted")
		return TorrentFile{}, err
	}

	numHashes := len(piece) / leng
	hashes := make([][20]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], piece[i*leng:(i+1)*leng])
	}

	t := TorrentFile{
		Announce:    b.Announce,
		InfoHash:    h,
		PieceHashes: hashes,
		PieceLength: b.Info.PieceLength,
		Length:      b.Info.Length,
		Name:        b.Info.Name,
	}

	return t, nil
}
