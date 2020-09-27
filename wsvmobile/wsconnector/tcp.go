package wsconnector

import (
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/gorilla/websocket"
	"io"
	"net"
)

func (t WsConnector) Handle(tcp net.Conn, target *net.TCPAddr) error {
	log.Infof("tcp connect %v\n", target.String())

	t.klist.AddKiller(tcp, func() {
		e := tcp.Close()
		if e != nil {
			log.Errorf("failed to close tcp ws %v", e)
		}
	})

	ws, _, e := websocket.DefaultDialer.Dial(t.server, nil)
	if e != nil {
		log.Errorf("failed do dial %v", e)
		t.klist.Kill(tcp)
		return e
	}

	t.klist.AddKiller(ws, func() {
		e = ws.Close()
		if e != nil {
			log.Errorf("failed to close tcp tcp %v", e)
		}
	})

	m := &msg{
		ConType: "tcp",
		Dst:     target.String(),
	}
	e = ws.WriteJSON(m)
	if e != nil {
		log.Errorf("failed do send init tcp connection %v", e)
		t.klist.Kill(tcp)
		t.klist.Kill(ws)
		return e
	}

	//ws->tcp
	go func() {
		defer t.klist.Kill(tcp)
		defer t.klist.Kill(ws)

		buf := make([]byte, 32*1024)
		for {
			_, r, e := ws.NextReader()
			if e != nil {
				if e, ok := e.(*websocket.CloseError); ok {
					log.Debugf("tcp ws connection closed")
				} else {
					log.Errorf("failed read ws tcp %v", e)
				}
				return
			}

			_, e = io.CopyBuffer(tcp, r, buf)
			if e != nil {
				log.Errorf("failed write tcp tcp %v", e)
				return
			}
		}
	}()

	//tcp->ws
	go func() {
		defer t.klist.Kill(tcp)
		defer t.klist.Kill(ws)

		buf := make([]byte, 32*1024)
		for {
			n, e := tcp.Read(buf)
			if e != nil {
				if e == io.EOF {
					return //tunnel close
				} else {
					log.Errorf("failed read tcp %v", e)
					return
				}
			}

			e = ws.WriteMessage(websocket.BinaryMessage, buf[:n])
			if e != nil {
				log.Errorf("failed do write ws %v", e)
				return
			}
		}
	}()

	return nil
}
