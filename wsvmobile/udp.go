package wsvmobile

import (
	"errors"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/gorilla/websocket"
	"net"
)

// Connect connects the proxy server. Note that target can be nil.
func (t tcpUdpToWs) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	c, _, e := websocket.DefaultDialer.Dial(t.server, nil)
	if e != nil {
		log.Errorf("failed do dial %v", e)
		return e
	}

	m := &msg{
		ConType: "udp",
		Dst:     conn.LocalAddr().String(),
		Port:    conn.LocalAddr().Port,
	}
	e = c.WriteJSON(m)
	if e != nil {
		log.Errorf("failed do init udp ws %v", e)
		return e
	}
	t.clients[conn] = c

	go func() {
		defer func() {
			e := c.Close()
			if e != nil {
				log.Errorf("failed to close udp ws %v", e)
			}
			c = nil
		}()

		for {
			tp, b, e := c.ReadMessage()
			if e != nil {
				log.Errorf("failed do read from udp %v", e)
				return
			}

			if tp != websocket.BinaryMessage {
				log.Errorf("unexpected udp message type from ws")
				return
			}

			for len(b) > 0 {
				n, e := conn.WriteFrom(b, target)
				if e != nil {
					log.Errorf("failed to write to tun from udp ws %v", e)
					return
				}
				b = b[n:]
			}
		}
	}()
	return nil
}

// ReceiveTo will be called when data arrives from TUN.
func (t tcpUdpToWs) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	log.Infof("udp to %v", addr.String())
	ws := t.clients[conn]
	if ws == nil {
		return errors.New("ws doesn't exist for conn")
	}

	m := &msg{
		ConType: "udp",
		Dst:     addr.String(),
		Port:    addr.Port,
	}
	e := ws.WriteJSON(m)
	if e != nil {
		log.Errorf("failed do send udp header to ws %v", e)
		return e
	}

	e = ws.WriteMessage(websocket.BinaryMessage, data)
	if e != nil {
		log.Errorf("failed do send udp to ws %v", e)
		return e
	}
	return nil
}
