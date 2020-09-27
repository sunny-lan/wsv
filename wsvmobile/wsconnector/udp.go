package wsconnector

import (
	"errors"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/gorilla/websocket"
	"net"
	"sync"
)

// Connect connects the proxy server. Note that target can be nil.
func (t WsConnector) Connect(udp core.UDPConn, target *net.UDPAddr) error {
	log.Debugf("UDP connect %v", target.String())

	t.klist.AddKiller(udp, func() {
		t.udpWs.Delete(udp)

		e := udp.Close()
		if e != nil {
			log.Errorf("failed to close udp tun %v", e)
		}
	})

	ws, _, e := websocket.DefaultDialer.Dial(t.server, nil)
	if e != nil {
		log.Errorf("failed do dial udp ws %v", e)
		t.klist.Kill(udp)
		return e
	}

	lock := &sync.Mutex{}

	t.klist.AddKiller(ws, func() {
		lock.Lock()
		e := ws.Close()
		lock.Unlock()
		if e != nil {
			log.Errorf("failed to close udp ws %v", e)
		}
	})

	m := &msg{
		ConType: "udp",
		Dst:     udp.LocalAddr().String(),
		Port:    udp.LocalAddr().Port,
	}
	lock.Lock()
	e = ws.WriteJSON(m)
	lock.Unlock()
	if e != nil {
		log.Errorf("failed do init udp ws %v", e)
		t.klist.Kill(udp)
		t.klist.Kill(ws)
		return e
	}

	t.udpWs.Store(udp, lockedWs{
		ws:   ws,
		lock: lock,
	})

	go func() {
		defer t.klist.Kill(ws)
		defer t.klist.Kill(udp)

		for {
			tp, b, e := ws.ReadMessage()
			if e != nil {
				if e, ok := e.(*websocket.CloseError); ok {
					log.Debugf("udp ws connection closed")
				} else {
					log.Errorf("failed to read message from udp ws %v", e)
				}
				return
			}

			if tp != websocket.BinaryMessage {
				log.Errorf("unexpected udp message type from ws")
				return
			}

			for len(b) > 0 {
				n, e := udp.WriteFrom(b, target)
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
func (t WsConnector) ReceiveTo(udp core.UDPConn, data []byte, addr *net.UDPAddr) error {
	log.Debugf("udp to %v", addr.String())
	v, ok := t.udpWs.Load(udp)
	if !ok {
		return errors.New("lws doesn't exist for udp")
	}

	lws := v.(lockedWs)

	m := &msg{
		ConType: "udp",
		Dst:     addr.String(),
		Port:    addr.Port,
	}
	lws.lock.Lock()
	e := lws.ws.WriteJSON(m)
	lws.lock.Unlock()
	if e != nil {
		log.Errorf("failed do send udp header to lws %v", e)
		return e
	}

	lws.lock.Lock()
	e = lws.ws.WriteMessage(websocket.BinaryMessage, data)
	lws.lock.Unlock()
	if e != nil {
		log.Errorf("failed do send udp to lws %v", e)
		return e
	}

	return nil
}
