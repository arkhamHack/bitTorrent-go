package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

//will tel if peer is avialable or not(choked or unchoked)
type msgId uint8

const (
	MsgChoke         msgId = 0
	MsgUnchoke       msgId = 1
	MsgInterested    msgId = 2
	MsgNotinterested msgId = 3
	MsgHave          msgId = 4
	MsgBitfield      msgId = 5
	MsgRequest       msgId = 6
	MsgPiece         msgId = 7
	MsgCancel        msgId = 8
)

type Msg struct {
	Id      msgId
	Payload []byte
}

func (m *Msg) SeriaLize() []byte {

	if m == nil {
		buffer := make([]byte, 4) //length indicator is 4 bytes
		return buffer
	}
	buffer := make([]byte, 4+(uint32(len(m.Payload)+1)))
	binary.BigEndian.PutUint32(buffer, uint32(len(m.Payload)+1))
	buffer[4] = byte(m.Id)
	copy(buffer[5:], m.Payload)

	return buffer
}

//handles REQUEST id message
// func reQ(index, begin, length int) *msg {
// 	payload := make([]byte, 12)

// }

//handles HAVE id message

//handles
//parsing the message from the stream
func Read(r io.Reader) (*Msg, error) {
	bufferLength := make([]byte, 4)
	_, err := io.ReadFull(r, bufferLength)
	if err != nil {
		return nil, err
	}
	sizeBin := binary.BigEndian.Uint32(bufferLength)

	if sizeBin == 0 {
		return nil, nil
	}
	msgBuf := make([]byte, sizeBin)
	_, err = io.ReadFull(r, msgBuf)
	if err != nil {
		return nil, err
	}
	m := Msg{
		Id:      msgId(msgBuf[0]),
		Payload: msgBuf[1:],
	}
	return &m, nil
}

func (m *Msg) name() string {
	if m == nil {
		return "KEEPALIVE"
	}
	switch m.Id {
	case MsgChoke:
		return "CHOKE"

	case MsgUnchoke:
		return "UNCHOKE"

	case MsgHave:
		return "HAVE"

	case MsgInterested:
		return "INTERESTED"
	case MsgNotinterested:
		return "NOTINTERESTED"

	case MsgBitfield:
		return "BITFIELD"

	case MsgPiece:
		return "PIECE"

	case MsgRequest:
		return "REQUEST"

	case MsgCancel:
		return "CANCEL"
	default:
		return fmt.Sprintf("unknown: %d", m.Id)
	}

}
func FormatRequest(index, begin, len int) *Msg {

	pd := make([]byte, 12)
	binary.BigEndian.PutUint32(pd[0:4], uint32(index))
	binary.BigEndian.PutUint32(pd[4:8], uint32(begin))
	binary.BigEndian.PutUint32(pd[8:12], uint32(len))

	return &Msg{Id: MsgRequest, Payload: pd}
}
func FormatHave(index int) *Msg {

	pd := make([]byte, 4)
	binary.BigEndian.PutUint32(pd, uint32(index))

	return &Msg{Id: MsgHave, Payload: pd}
}
func (m *Msg) String() string {
	if m == nil {
		return m.name()
	}

	return fmt.Sprintf("%s with size: %d", m.name(), len(m.Payload))
}

func ParseHave(msg *Msg) (int, error) {
	if msg.Id != MsgHave {
		return 0, fmt.Errorf("HAVE expected instead received %d", msg.Id)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("Payload length expected 4, got %d", len(msg.Payload))
	}
	n := int(binary.BigEndian.Uint32(msg.Payload))
	return n, nil
}

func ParseParts(index int, buf []byte, msg *Msg) (int, error) {
	if msg.Id != MsgPiece {
		return 0, fmt.Errorf("PIECE expected instead received %d", msg.Id)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("Payload length expected 8, got %d", len(msg.Payload))
	}
	prIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if prIndex != index {
		return 0, fmt.Errorf("Index expected %d not %d", index, prIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin offset too far off %d", begin)
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("Begin offset too far off %d", begin)
	}
	//copy(buf[begin:], data)
	return len(data), nil
}
