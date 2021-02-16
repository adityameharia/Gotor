package file

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

func (t *TorrentFile) requestPeers(Pid []byte, port uint16) error {
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
		return err
	}
	defer resp.Body.Close()

	tracker := Tracker{}
	err = bencode.Unmarshal(resp.Body, &tracker)
	if err != nil {
		return err
	}

	DecodePeer([]byte(tracker.Peers))

	return nil
}

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
