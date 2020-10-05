package wsconnector

import (
	"errors"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"sync"
)

type udpWrap struct {
	udp    core.UDPConn
	target *net.UDPAddr
}

func (u udpWrap) Write(p []byte) (n int, err error) {
	return u.udp.WriteFrom(p, u.target)
}

// Connect connects the proxy originalServer. Note that target can be nil.
func (t WsConnector) Connect(udp core.UDPConn, target *net.UDPAddr) error {
	log.Debugf("UDP connect %v", target.String())

	t.Stats.udpStats.AddConnection()

	t.kList.AddKiller(udp, func() {
		t.udpWs.Delete(udp)

		e := udp.Close()
		if e != nil {
			log.Errorf("failed to close udp tun %v", e)
		}

		t.Stats.udpStats.RemoveConnection()
	})

	t.Stats.wsStats.AddConnection()

	ws, _, e := t.dialWs()
	if e != nil {
		log.Errorf("failed do dial udp ws %v", e)
		t.kList.Kill(udp)
		return e
	}

	lock := &sync.Mutex{}

	t.kList.AddKiller(ws, func() {
		lock.Lock()
		e := ws.Close()
		lock.Unlock()
		if e != nil {
			log.Errorf("failed to close udp ws %v", e)
		}

		t.Stats.wsStats.RemoveConnection()
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
		t.kList.Kill(udp)
		t.kList.Kill(ws)
		return e
	}

	t.udpWs.Store(udp, lockedWs{
		ws:   ws,
		lock: lock,
	})

	go func() {
		defer t.kList.Kill(ws)
		defer t.kList.Kill(udp)

		wrapped := udpWrap{
			udp:    udp,
			target: target,
		}
		buf := make([]byte, 1024*32)
		for {
			tp, r, e := ws.NextReader()
			if e != nil {
				if _, ok := e.(*websocket.CloseError); ok {
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
			n, e := io.CopyBuffer(wrapped, r, buf)
			t.Stats.wsStats.RecordRead(n)
			t.Stats.udpStats.RecordWrite(n)
			if e != nil {
				log.Errorf("failed to write to tun from udp ws %v", e)
				return
			}
		}
	}()
	return nil
}

// ReceiveTo will be called when data arrives from TUN.
func (t WsConnector) ReceiveTo(udp core.UDPConn, data []byte, addr *net.UDPAddr) error {
	t.Stats.udpStats.RecordRead(int64(len(data)))

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
	defer lws.lock.Unlock()

	e := lws.ws.WriteJSON(m)
	if e != nil {
		log.Errorf("failed do send udp header to lws %v", e)
		return e
	}

	e = lws.ws.WriteMessage(websocket.BinaryMessage, data)
	if e != nil {
		log.Errorf("failed do send udp to lws %v", e)
		return e
	}

	t.Stats.wsStats.RecordWrite(int64(len(data)))

	return nil
}
