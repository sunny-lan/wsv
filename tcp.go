package wsvmobile

import (
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/gorilla/websocket"
	"net"
	"runtime/debug"
)

type socktosock struct {
	server  string
	udpWs   *websocket.Conn
	clients map[int]core.UDPConn
	addrs   map[int]*net.UDPAddr
}

type msg struct {
	conType string
	dst     string
	port    int
}

func makess(server string) *socktosock {
	return &socktosock{
		server,
		nil,
		make(map[int]core.UDPConn),
		make(map[int]*net.UDPAddr),
	}
}

func (t socktosock) getUdpWs() (*websocket.Conn, error) {
	if t.udpWs == nil {
		c, _, e := websocket.DefaultDialer.Dial(t.server, nil)
		if e != nil {
			log.Fatalf("failed do dial %v", e)
			return nil, e
		}

		t.udpWs = c

		m := msg{
			conType: "udp",
			dst:     "",
			port:    -1,
		}
		e = c.WriteJSON(m)
		if e != nil {
			log.Fatalf("failed do init udp ws %v", e)
			return nil, e
		}

		go func() {
			defer func() {
				t.udpWs.Close()
				t.udpWs = nil
			}()

			for {
				var m msg
				e := t.udpWs.ReadJSON(m)
				if e != nil {
					log.Fatalf("failed do read udp header from ws %v", e)
					return
				}
				if m.conType != "udp" {
					log.Fatalf("recieved non udp message from ws")
					return
				}

				tp, b, e := t.udpWs.ReadMessage()
				if e != nil {
					log.Fatalf("failed do read from udp %v", e)
					return
				}

				if tp != websocket.BinaryMessage {
					log.Fatalf("unexpected udp message type from ws")
					return
				}

				client, ok := t.clients[m.port]
				if ok {
					client.WriteFrom(b, t.addrs[m.port])
				}
			}

		}()
	}
	return t.udpWs, nil
}

// Connect connects the proxy server. Note that target can be nil.
func (t socktosock) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	_, e := t.getUdpWs()
	if e != nil {
		return e
	}
	t.clients[target.Port] = conn
	t.addrs[target.Port] = target
	return nil
}

// ReceiveTo will be called when data arrives from TUN.
func (t socktosock) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	ws, e := t.getUdpWs()
	if e != nil {
		return e
	}

	m := msg{
		conType: "udp",
		dst:     addr.String(),
		port:    addr.Port,
	}
	e = ws.WriteJSON(m)
	if e != nil {
		log.Fatalf("failed do send udp header to ws %v", e)
		return e
	}

	e = ws.WriteMessage(websocket.BinaryMessage, data)
	if e != nil {
		log.Fatalf("failed do send udp to ws %v", e)
		return e
	}
	return nil
}

func (t socktosock) Handle(tun net.Conn, target *net.TCPAddr) error {
	log.Infof("tcp to %v\n", target.String())
	debug.PrintStack()

	c, _, e := websocket.DefaultDialer.Dial(t.server, nil)
	if e != nil {
		log.Fatalf("failed do dial %v", e)
		return e
	}

	m := msg{
		conType: "tcp",
		dst:     target.String(),
	}
	e = c.WriteJSON(&m)

	var killed bool = false

	var kill = func() {
		if killed {
			return
		}
		killed = true
	}
	//ws->tun
	go func() {
		defer kill()
		for {
			_, b, e := c.ReadMessage()
			if e != nil {
				log.Fatalf("failed do read ws %v", e)
				return
			}
			for len(b) > 0 {
				n, e := tun.Write(b)
				if e != nil {
					log.Fatalf("failed do write tun %v", e)
					return
				}
				b = b[n:]
			}
		}
	}()

	//tun->ws
	go func() {
		defer kill()
		buf := make([]byte, 1024)
		for {
			n, e := tun.Read(buf)
			if e != nil {
				log.Fatalf("failed read tun %v", e)
				return
			}
			e = c.WriteMessage(websocket.BinaryMessage, buf[:n])
			if e != nil {
				log.Fatalf("failed do write ws %v", e)
				return
			}
		}
	}()

	return nil
}
