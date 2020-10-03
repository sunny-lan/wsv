package wsconnector

import (
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/gorilla/websocket"
	"github.com/sunny-lan/wsv/common"
	"io"
	"net"
)

func (t WsConnector) Handle(tcp net.Conn, target *net.TCPAddr) error {
	log.Infof("tcp connect %v\n", target.String())

	t.Stats.TcpStats.AddConnection()

	t.klist.AddKiller(tcp, func() {
		e := tcp.Close()
		if e != nil {
			log.Errorf("failed to close tcp ws %v", e)
		}

		t.Stats.TcpStats.RemoveConnection()
	})

	t.Stats.WsStats.AddConnection()
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

		t.Stats.WsStats.RemoveConnection()
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
			nr, r, e := ws.NextReader()
			t.Stats.WsStats.Update(func(stats *common.ConnStats) {
				stats.ReadLoops++
				stats.BytesRead += int64(nr)
			})
			if e != nil {
				if e, ok := e.(*websocket.CloseError); ok {
					log.Debugf("tcp ws connection closed")
				} else {
					log.Errorf("failed read ws tcp %v", e)
				}
				return
			}

			nw, e := io.CopyBuffer(tcp, r, buf)
			t.Stats.TcpStats.Update(func(stats *common.ConnStats) {
				stats.WriteLoops++
				stats.BytesWritten += nw
			})
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
			t.Stats.TcpStats.Update(func(stats *common.ConnStats) {
				stats.ReadLoops++
				stats.BytesRead += int64(n)
			})
			if e != nil {
				if e == io.EOF {
					return //tunnel close
				} else {
					log.Errorf("failed read tcp %v", e)
					return
				}
			}

			e = ws.WriteMessage(websocket.BinaryMessage, buf[:n])
			t.Stats.WsStats.Update(func(stats *common.ConnStats) {
				stats.WriteLoops++
				stats.BytesWritten += int64(n)
			})
			if e != nil {
				log.Errorf("failed do write ws %v", e)
				return
			}

		}
	}()

	return nil
}
