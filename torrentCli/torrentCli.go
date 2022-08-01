package torrentCli

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/arkhamHack/bitTorrent-go/p2p"
	"github.com/jackpal/bencode-go"
)

type BencodeInfo struct {
	Part       string
	PartLength int
	Name       string
	Size       int
}

type BencodeTorrent struct {
	Announce string
	Info     BencodeInfo
}
type TorrentFile struct {
	Announce string
	InfoHash [20]byte
	PartHash [][20]byte
	PartLen  int
	Len      int
	Name     string
}

const Port = 6881

func (info *BencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, info)
	if err != nil {
		return [20]byte{}, err
	}
	hash := sha1.Sum(buf.Bytes())
	return hash, nil
}

func (bt BencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infohash, err := bt.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}
	pieceHashes, err := bt.Info.splitPieceHash()
	if err != nil {
		return TorrentFile{}, err
	}
	t := TorrentFile{
		Announce: bt.Announce,
		InfoHash: infohash,
		PartHash: pieceHashes,
		PartLen:  bt.Info.PartLength,
		Len:      bt.Info.Size,
		Name:     bt.Info.Name,
	}
	return t, nil
}
func (info *BencodeInfo) splitPieceHash() ([][20]byte, error) {
	hLen := 20 //length of sha-1 hash
	buf := []byte(info.Part)
	if len(buf)%hLen != 0 {
		err := fmt.Errorf("hash length is incorrect.Not proper hash")
		return nil, err
	}
	numOfhash := len(buf) / hLen
	hashes := make([][20]byte, numOfhash)

	for i := 0; i < numOfhash; i++ {
		copy(hashes[i][:], buf[i*hLen:(i+1)*hLen])
	}
	return hashes, nil
}

func (t *TorrentFile) DownloadToFile(path string) error {
	var pid [20]byte
	_, err := rand.Read(pid[:])
	if err != nil {
		return err
	}
	peers, err := t.requestPeers(pid, Port)
	if err != nil {
		return err
	}
	tor := p2p.Torrent{
		Peers:       peers,
		PieceLength: t.PartLen,
		PeerId:      pid,
		Len:         t.Len,
		Infohash:    t.InfoHash,
		PieceHashes: t.PartHash,
		Name:        t.Name,
	}
	buf, err := tor.Download()
	if err != nil {
		return err
	}
	outputFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	_, err = outputFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	bt := BencodeTorrent{}
	err = bencode.Unmarshal(file, &bt)
	if err != nil {
		return TorrentFile{}, err
	}
	return bt.toTorrentFile()

}
