package wsconnector

import (
	"github.com/sunny-lan/wsv/common"
)

type WsConnStatsReader interface {
	UDP() common.ConnStatsReader
	TCP() common.ConnStatsReader
	WS() common.ConnStatsReader
	ResetCounters()
}

type wsConnStats struct {
	udpStats *common.ConnStats
	tcpStats *common.ConnStats
	wsStats  *common.ConnStats
}

func (s *wsConnStats) UDP() common.ConnStatsReader {
	return s.udpStats
}

func (s *wsConnStats) TCP() common.ConnStatsReader {
	return s.tcpStats
}

func (s *wsConnStats) WS() common.ConnStatsReader {
	return s.wsStats
}

func newWsConnStats() *wsConnStats {
	return &wsConnStats{
		common.NewConnStats(),
		common.NewConnStats(),
		common.NewConnStats(),
	}
}

func (s *wsConnStats) ResetCounters() {
	s.udpStats.ResetCounters()
	s.tcpStats.ResetCounters()
	s.wsStats.ResetCounters()
}
