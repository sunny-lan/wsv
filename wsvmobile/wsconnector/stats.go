package wsconnector

import (
	"github.com/sunny-lan/wsv/common"
)

type WsConnStats struct {
	UdpStats *common.ConnStats
	TcpStats *common.ConnStats
	WsStats  *common.ConnStats
}

func NewWsConnStats() *WsConnStats {
	return &WsConnStats{
		common.NewConnStats(),
		common.NewConnStats(),
		common.NewConnStats(),
	}
}

func (s WsConnStats) ResetCounters() {
	s.UdpStats.ResetCounters()
	s.TcpStats.ResetCounters()
	s.WsStats.ResetCounters()
}
