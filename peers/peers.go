package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Peer struct {
	Ip   net.IP
	Port uint16
}

func UnmarshalPeer(peerList []byte) ([]Peer, error) {
	const peerSize = 6
	peerNum := len(peerList) / peerSize
	if (peerNum)%peerSize != 0 {
		err := fmt.Errorf("Peers are malfunctioning")
		return nil, err
	}
	peers := make([]Peer, peerNum)
	for i := 0; i < peerNum; i++ {
		offSet := i * peerSize
		peers[i].Ip = net.IP(peerList[offSet : offSet+4])
		peers[i].Port = binary.BigEndian.Uint16(peerList[offSet+4 : offSet+6])

	}
	return peers, nil
}

func (p Peer) Connect() string {
	return net.JoinHostPort(p.Ip.String(), strconv.Itoa(int(p.Port)))
}
