package protocol

import (
	"github.com/sunny-lan/wsv/wsvmobile/connection"
	"io"
	"net"
)

type TCPConn io.ReadWriteCloser
type UDPListener func(data []byte, addr *net.UDPAddr) error
type UDPConn interface {
	io.Writer
	SetListener(listener UDPListener)
}

type Protocol interface {
	HandleTCP(t TCPConn, target *net.TCPAddr) error
	HandleUDP(u UDPConn, target *net.UDPAddr) error
}

type Constructor interface {
	Make(dialer connection.Dialer) Protocol
}
