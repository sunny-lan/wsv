package wsvmobile

import (
	"github.com/sunny-lan/wsv/common"
	"github.com/sunny-lan/wsv/wsvmobile/wsconnector"
)

//TODO move into common, StatReader+Stat
type ConnStats interface {
	Connections() int64
	NewConnections() int64
	ReadLoops() int64
	WriteLoops() int64
	BytesRead() int64
	BytesWritten() int64
	Clear()
}

type WsvStats interface {
	UDP() ConnStats
	TCP() ConnStats
	Ws() ConnStats
}

type wsConnStatsWrapper struct {
	udp *connStatWrapper
	tcp *connStatWrapper
	ws  *connStatWrapper
}

func (w *wsConnStatsWrapper) UDP() ConnStats {
	return w.udp
}

func (w *wsConnStatsWrapper) TCP() ConnStats {
	return w.tcp
}

func (w *wsConnStatsWrapper) Ws() ConnStats {
	return w.ws
}

func newWsConnStatsWrapper(stat *wsconnector.WsConnStats) *wsConnStatsWrapper {
	return &wsConnStatsWrapper{
		udp: newConnStatWrapper(stat.UdpStats),
		tcp: newConnStatWrapper(stat.TcpStats),
		ws:  newConnStatWrapper(stat.WsStats),
	}
}

type connStatWrapper struct {
	stat *common.ConnStats
}

func newConnStatWrapper(stat *common.ConnStats) *connStatWrapper {
	return &connStatWrapper{
		stat: stat,
	}
}

func (c *connStatWrapper) Connections() int64 {
	c.stat.Lock.Lock()
	defer c.stat.Lock.Unlock()
	return c.stat.Connections
}

func (c *connStatWrapper) NewConnections() int64 {
	c.stat.Lock.Lock()
	defer c.stat.Lock.Unlock()
	return c.stat.NewConnections
}

func (c *connStatWrapper) ReadLoops() int64 {
	c.stat.Lock.Lock()
	defer c.stat.Lock.Unlock()
	return c.stat.ReadLoops
}

func (c *connStatWrapper) WriteLoops() int64 {
	c.stat.Lock.Lock()
	defer c.stat.Lock.Unlock()
	return c.stat.WriteLoops
}

func (c *connStatWrapper) BytesRead() int64 {
	c.stat.Lock.Lock()
	defer c.stat.Lock.Unlock()
	return c.stat.BytesRead
}

func (c *connStatWrapper) BytesWritten() int64 {
	c.stat.Lock.Lock()
	defer c.stat.Lock.Unlock()
	return c.stat.BytesWritten
}

func (c *connStatWrapper) Clear() {
	c.stat.ResetCounters()
}
