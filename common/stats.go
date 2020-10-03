package common

import "sync"

type ConnStatsReader interface {
	Connections() int64
	NewConnections() int64
	ReadLoops() int64
	WriteLoops() int64
	BytesRead() int64
	BytesWritten() int64
	ResetCounters()
}

type ConnStats struct {
	Connections    int64
	NewConnections int64
	ReadLoops      int64
	WriteLoops     int64
	BytesRead      int64
	BytesWritten   int64
	Lock           *sync.Mutex
}

func NewConnStats() *ConnStats {
	return &ConnStats{
		Lock: &sync.Mutex{},
	}
}

func (s *ConnStats) ResetCounters() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.NewConnections = 0
	s.ReadLoops = 0
	s.WriteLoops = 0
	s.BytesRead = 0
	s.BytesWritten = 0
}

func (s *ConnStats) AddConnection() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.Connections++
	s.NewConnections++
}

func (s *ConnStats) RemoveConnection() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.Connections--
}

func (s *ConnStats) Update(update func(stats *ConnStats)) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	update(s)
}

//type ErrorSeverity int
//
//const (
//	//FATAL means the VPN cannot continue running
//	FATAL ErrorSeverity = iota
//	WARNING
//
//)
//
//type ErrorSource int
//
//const (
//	//FATAL means the VPN cannot continue running
//	UDP ErrorSource = iota
//	TCP
//
//)
//
//
//type InfoListener interface {
//	OnError(e error)
//}
