package connector

import (
	"encoding/json"
	"github.com/eycorsican/go-tun2socks/common/log"
	"io"
	"net"
)

func (t *Connector) Handle(tcp net.Conn, target *net.TCPAddr) error {
	log.Infof("tcp connect %v\n", target.String())
	successStart := false

	t.Stats.tcpStats.AddConnection()

	t.kList.AddKiller(tcp, func() {
		e := tcp.Close()
		if e != nil {
			log.Errorf("failed to close tcp ws %v", e)
		}

		t.Stats.tcpStats.RemoveConnection()
	})
	defer func() {
		if !successStart {
			t.kList.Kill(tcp)
		}
	}()

	t.Stats.wsStats.AddConnection()
	ws, e := t.dialer.Dial()
	if e != nil {
		log.Errorf("failed do dial %v", e)
		return e
	}

	t.kList.AddKiller(ws, func() {
		e = ws.Close()
		if e != nil {
			log.Errorf("failed to close tcp tcp %v", e)
		}

		t.Stats.wsStats.RemoveConnection()
	})
	defer func() {
		if !successStart {
			t.kList.Kill(ws)
		}
	}()

	m := &msg{
		ConType: "tcp",
		Dst:     target.String(),
	}
	b, e := json.Marshal(m)
	if e != nil {
		log.Errorf("failed to ")
		return e
	}

	_, e = ws.Write(b)
	if e != nil {
		log.Errorf("failed do send init tcp connection %v", e)
		return e
	}

	//ws->tcp
	go func() {
		defer t.kList.Kill(tcp)
		defer t.kList.Kill(ws)

		buf := make([]byte, 32*1024)
		_, e := io.CopyBuffer(ws, tcp, buf)
		if e != nil {
			log.Errorf("ws->tcp loop failed %v", e)
		}
	}()

	//tcp->ws
	go func() {
		defer t.kList.Kill(tcp)
		defer t.kList.Kill(ws)

		buf := make([]byte, 32*1024)
		_, e := io.CopyBuffer(tcp, ws, buf)
		if e != nil {
			log.Errorf("tcp->ws loop failed %v", e)
		}
	}()

	successStart = true

	return nil
}
