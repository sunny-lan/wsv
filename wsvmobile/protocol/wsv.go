package protocol

import (
	"net"
)

type Wsv struct {
}

func (w *Wsv) HandleTCP(t TCPConn, target *net.TCPAddr) error {
	panic("implement me")
}

func (w *Wsv) HandleUDP(u UDPConn, target *net.UDPAddr) error {
	panic("implement me")
}
