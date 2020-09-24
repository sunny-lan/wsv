package wsvmobile

import (
	"errors"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/eycorsican/go-tun2socks/common/log/simple"
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/gorilla/websocket"
	"github.com/higebu/netfd"
	"io"
	"net"
	"os"
)

type Protector interface {
	Protect(fd int)
}

type PacketFlow interface {
	WritePacket(packet []byte) (int, error)
}

var lwipStack core.LWIPStack

func InputPacket(data []byte, offset int, length int) (int, error) {
	return lwipStack.Write(data[offset : offset+length])
}

func ConnectRaw(packetFlow PacketFlow, proxyHost string, protector Protector) {
	start(proxyHost, protector, packetFlow.WritePacket)
}

func start(proxyHost string, protector Protector, fn func([]byte) (int, error)) {
	log.RegisterLogger(simple.NewSimpleLogger())
	log.SetLevel(log.INFO)

	if protector != nil {
		websocket.DefaultDialer.NetDial = func(network, addr string) (net.Conn, error) {
			c, e := net.Dial(network, addr)
			if e != nil {
				log.Fatalf("unable to establish connection %v", e)
				return nil, e
			}
			protector.Protect(netfd.GetFdFromConn(c))
			return c, nil
		}
	}

	var s = makess(proxyHost)
	lwipStack = core.NewLWIPStack()
	core.RegisterTCPConnHandler(s)
	core.RegisterUDPConnHandler(s)
	core.RegisterOutputFn(fn)
}

func ConnectFd(tunfd int64, proxyHost string, protector Protector) error {
	f := os.NewFile(uintptr(tunfd), "tunfd")
	defer func() {
		e := f.Close()
		if e != nil {
			log.Fatalf("tunfd already closed")
		}
	}()
	if f == nil {
		log.Fatalf("invalid tunfd")
		return errors.New("invalid tunfd")
	}
	start(proxyHost, protector, f.Write)
	_, e := io.Copy(lwipStack, f)
	return e
}
