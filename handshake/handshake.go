package handshake

import (
	"fmt"
	"io"
)

type Handshake struct {
	proto_str string
	InfoHash  [20]byte
	peerId    [20]byte
}

func New(infohash, peerId [20]byte) *Handshake {
	return &Handshake{
		proto_str: "BitTorrent protocol",
		InfoHash:  infohash,
		peerId:    peerId,
	}
}
func (h *Handshake) SeriaLize() []byte {

	buffer := make([]byte, len(h.proto_str)+49)
	buffer[0] = byte(len(h.proto_str))
	copy(buffer[1:], h.proto_str)
	curr := copy(buffer[1+len(h.proto_str):], make([]byte, 8)) //reserving 8 bytes for extensions
	curr += copy(buffer[curr:], h.InfoHash[:])
	curr += copy(buffer[curr:], h.peerId[:])
	return buffer
}

func Read(r io.Reader) (*Handshake, error) {
	bufferLength := make([]byte, 1)
	_, err := io.ReadFull(r, bufferLength)
	if err != nil {

		return nil, err

	}
	if int(bufferLength[0]) == 0 {
		err := fmt.Errorf("protocol string can't be 0 length")
		return nil, err
	}
	prLen := int(bufferLength[0])
	buffer := make([]byte, 48+prLen)
	_, err = io.ReadFull(r, buffer)
	if err != nil {
		return nil, err
	}

	var info, pd [20]byte
	copy(info[:], buffer[prLen+8:prLen+28])
	copy(pd[:], buffer[prLen+28:])

	hand := Handshake{
		proto_str: string(buffer[0:prLen]),
		InfoHash:  info,
		peerId:    pd,
	}
	return &hand, nil
}
