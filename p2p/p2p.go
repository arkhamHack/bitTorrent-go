package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/arkhamHack/bitTorrent-go/client"
	"github.com/arkhamHack/bitTorrent-go/message"
	"github.com/arkhamHack/bitTorrent-go/peers"
)

const MaxBlockSize = 16384 //max number of bytes a request can ask for
const MaxBacklog = 5       //max number of unfulfilled requests a client can have in its pipeline
type Torrent struct {
	Peers       []peers.Peer
	PieceLength int
	PeerId      [20]byte
	Len         int
	Infohash    [20]byte
	PieceHashes [][20]byte
	Name        string
}
type piece struct {
	index  int
	hash   [20]byte
	length int
}

type pieceRes struct {
	index int
	buf   []byte
}

type partProg struct {
	index    int
	download int
	req      int
	backlog  int
	client   *client.Client
	buf      []byte
}

func (p *partProg) readMsg() error {
	msg, err := p.client.Read()
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}
	switch msg.Id {
	case message.MsgUnchoke:
		p.client.Choked = false
	case message.MsgChoke:
		p.client.Choked = true

	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return nil
		}
		p.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		pie, err := message.ParseParts(p.index, p.buf, msg)
		if err != nil {
			return err
		}
		p.download += pie
		p.backlog--
	}
	return nil
}
func attemptPieceDownload(c *client.Client, p *piece) ([]byte, error) {
	state := partProg{
		index:  p.index,
		client: c,
		buf:    make([]byte, p.length),
	}
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) //disabling deadline
	if state.download < p.length {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.req < p.length {
				blkSize := MaxBlockSize
				if p.length-state.req < blkSize {
					blkSize = p.length - state.req
				}
				err := state.client.ReQuest(p.index, state.req, blkSize)
				if err != nil {

					return nil, err
				}
				state.req += blkSize
				state.backlog--
			}
		}
		err := state.readMsg()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}
func (t *Torrent) calcPieceBound(index int) (int, int) {
	begin := t.PieceLength * index
	end := begin + t.PieceLength
	if end > t.Len {
		end = t.Len
	}
	return begin, end
}
func (t *Torrent) calcPieceSize(index int) int {
	begin, end := t.calcPieceBound(index)
	return end - begin
}

func checkIntegrity(pw *piece, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}
func (t *Torrent) downloadBegin(peer peers.Peer, wrkQ chan *piece, res chan *pieceRes) {
	c, err := client.New(peer, t.PeerId, t.Infohash)
	if err != nil {
		log.Printf("\nHandshake with %s Failed.", peer.Ip)
		return
	}
	defer c.Conn.Close()
	c.Unchoke()
	c.Interested()

	for pw := range wrkQ {
		if !c.Bitfield.HasPiece(pw.index) {

			wrkQ <- pw //sends worker back to queue
			continue
		}
		buf, err := attemptPieceDownload(c, pw)
		if err != nil {
			log.Println("Ending...", err)
			wrkQ <- pw
			return
		}
		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			wrkQ <- pw
			continue
		}
		c.Have(pw.index)
		res <- &pieceRes{pw.index, buf}
	}
}

//downloads torrent and saves file in memory

func (t *Torrent) Download() ([]byte, error) {
	log.Println("download of %s starting...", t.Name)
	wrkQ := make(chan *piece, len(t.PieceHashes))
	res := make(chan *pieceRes)
	for index, hash := range t.PieceHashes {
		len := t.calcPieceSize(index)
		wrkQ <- &piece{index, hash, len}
	}
	for _, peer := range t.Peers {
		go t.downloadBegin(peer, wrkQ, res)
	}

	buf := make([]byte, t.Len)
	completedPieces := 0
	for completedPieces < len(t.PieceHashes) {
		fin := <-res
		begin, end := t.calcPieceBound(fin.index)
		copy(buf[begin:end], fin.buf)
		prt := float64(completedPieces) / float64(len(t.PieceHashes)) * 100
		completedPieces++
		pid := runtime.NumGoroutine()
		log.Printf("%0.2f Downloaded #%d from peer %d \n", prt, fin.index, pid)
	}
	close(wrkQ)
	return buf, nil
}
