package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/arkhamHack/bitTorrent-go/bitfield"
	"github.com/arkhamHack/bitTorrent-go/handshake"
	"github.com/arkhamHack/bitTorrent-go/message"
	"github.com/arkhamHack/bitTorrent-go/peers"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	InfoHash [20]byte
	PeerId   [20]byte
	Bitfield bitfield.Bitfield
	Pr       peers.Peer
}

func CompHandshake(conn net.Conn, infohash, peerId [20]byte) (*handshake.Handshake, error) {

	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})
	req := handshake.New(infohash, peerId)
	_, err := conn.Write(req.SeriaLize())
	if err != nil {
		return nil, err
	}
	h, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(h.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("%x does not match the required %x", infohash, h.InfoHash)
	}

	return h, nil
}

func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("expected bitfield insted got: %s", msg)

		return nil, err
	}
	if msg.Id != message.MsgBitfield {
		err := fmt.Errorf("expected bitfield requested recieved id: %d", msg.Id)
		return nil, err
	}
	return msg.Payload, err

}

//connect to peer,receive handshake complete handshake
func New(peer peers.Peer, pid, infohash [20]byte) (*Client, error) {

	tcp, err := net.DialTimeout("tcp", peer.Connect(), 5*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = CompHandshake(tcp, infohash, pid)

	if err != nil {
		tcp.Close()
		return nil, err
	}

	bf, err := recvBitfield(tcp)
	if err != nil {
		tcp.Close()
		return nil, err
	}

	return &Client{
		Conn:     tcp,
		Choked:   true,
		InfoHash: infohash,
		Bitfield: bf,
		PeerId:   pid,
		Pr:       peer,
	}, nil
}

func (c *Client) Read() (*message.Msg, error) {
	msg, err := message.Read(c.Conn)

	return msg, err
}

func (c *Client) ReQuest(index, begin, len int) error {
	req := message.FormatRequest(index, begin, len)
	_, err := c.Conn.Write(req.SeriaLize())
	return err
}

func (c *Client) Interested() error {
	mg := message.Msg{Id: message.MsgInterested}

	_, err := c.Conn.Write(mg.SeriaLize())

	return err
}

func (c *Client) NotInterested() error {
	mg := message.Msg{Id: message.MsgNotinterested}

	_, err := c.Conn.Write(mg.SeriaLize())

	return err
}

func (c *Client) Unchoke() error {
	mg := message.Msg{Id: message.MsgUnchoke}

	_, err := c.Conn.Write(mg.SeriaLize())

	return err
}

func (c *Client) Have(index int) error {
	mg := message.FormatHave(index)

	_, err := c.Conn.Write(mg.SeriaLize())

	return err
}
