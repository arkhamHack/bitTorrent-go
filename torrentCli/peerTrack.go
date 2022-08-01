package torrentCli

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"

	"github.com/arkhamHack/bitTorrent-go/peers"
)

type tracker struct {
	Interval int
	Peers    string
}

func (t *TorrentFile) urlTracker(pid [20]byte, port uint16) (string, error) {
	Url, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	paras := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(pid[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"1"},
		"left":       []string{strconv.Itoa(t.Len)},
	}
	Url.RawQuery = paras.Encode()
	return Url.String(), nil

}
func (t *TorrentFile) requestPeers(pid [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := t.urlTracker(pid, port)
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Timeout: 20 * time.Second}
	res, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	trackRes := tracker{}
	err = bencode.Unmarshal(res.Body, trackRes)
	if err != nil {
		return nil, err
	}
	return peers.UnmarshalPeer([]byte(trackRes.Peers))
}
