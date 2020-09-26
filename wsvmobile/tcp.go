package wsvmobile

import (
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/gorilla/websocket"
	"io"
	"net"
)

type KillListener func() error

type tcpUdpToWs struct {
	server  string
	clients map[core.UDPConn]*websocket.Conn
}

type msg struct {
	ConType string `json:"ConType"`
	Dst     string `json:"Dst"`
	Port    int    `json:"Port"`
}

func newTcpUdpHandler(server string) *tcpUdpToWs {
	return &tcpUdpToWs{
		server,
		make(map[core.UDPConn]*websocket.Conn),
	}
}

func (t tcpUdpToWs) Handle(tun net.Conn, target *net.TCPAddr) error {
	log.Infof("tcp to %v\n", target.String())

	c, _, e := websocket.DefaultDialer.Dial(t.server, nil)
	if e != nil {
		log.Errorf("failed do dial %v", e)
		return e
	}

	m := &msg{
		ConType: "tcp",
		Dst:     target.String(),
	}
	e = c.WriteJSON(m)

	var killed bool = false

	var kill = func() {
		if killed {
			return
		}
		e := c.Close()
		if e != nil {
			log.Errorf("failed to close tcp ws %v", e)
			return
		}
		e = tun.Close()
		if e != nil {
			log.Errorf("failed to close tcp tun %v", e)
			return
		}
		killed = true
	}

	//ws->tun
	go func() {
		defer kill()
		buf := make([]byte, 32*1024)
		for {
			_, r, e := c.NextReader()
			if e != nil {
				log.Errorf("failed read ws tcp %v", e)
				return
			}
			n, e := io.CopyBuffer(tun, r, buf)
			if e != nil {
				log.Errorf("failed write tun tcp %v", e)
				return
			}
			log.Infof("wrote %v bytes from ws tcp to tun", n)
		}
	}()

	//tun->ws
	go func() {
		defer kill()
		buf := make([]byte, 32*1024)
		for {
			n, e := tun.Read(buf)
			if e != nil {
				if e == io.EOF {
					return //tunnel close
				} else {
					log.Errorf("failed read tun %v", e)
					return
				}
			}

			e = c.WriteMessage(websocket.BinaryMessage, buf[:n])
			if e != nil {
				log.Errorf("failed do write ws %v", e)
				return
			}
		}
	}()

	return nil
}
