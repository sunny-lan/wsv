package wsconnector

import (
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/gorilla/websocket"
	"io"
	"net"
)

func (t WsConnector) Handle(tcp net.Conn, target *net.TCPAddr) error {
	log.Infof("tcp connect %v\n", target.String())

	t.Stats.tcpStats.AddConnection()

	t.klist.AddKiller(tcp, func() {
		e := tcp.Close()
		if e != nil {
			log.Errorf("failed to close tcp ws %v", e)
		}

		t.Stats.tcpStats.RemoveConnection()
	})

	t.Stats.wsStats.AddConnection()
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

		t.Stats.wsStats.RemoveConnection()
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
			tp, r, e := ws.NextReader()
			if e != nil {
				if e, ok := e.(*websocket.CloseError); ok {
					log.Debugf("tcp ws connection closed")
				} else {
					log.Errorf("failed read ws tcp %v", e)
				}
				return
			}
			if tp != websocket.BinaryMessage {
				log.Errorf("unexpected message type from tcp ws")
				return
			}

			nw, e := io.CopyBuffer(tcp, r, buf)
			t.Stats.wsStats.RecordRead(nw)
			t.Stats.tcpStats.RecordWrite(nw)
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
			t.Stats.tcpStats.RecordRead(int64(n))
			if e != nil {
				if e == io.EOF {
					return //tunnel close
				} else {
					log.Errorf("failed read tcp %v", e)
					return
				}
			}

			e = ws.WriteMessage(websocket.BinaryMessage, buf[:n])
			t.Stats.wsStats.RecordWrite(int64(n))
			if e != nil {
				log.Errorf("failed do write ws %v", e)
				return
			}

		}
	}()

	return nil
}
