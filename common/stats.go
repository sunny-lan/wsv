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

type ConnStatsWriter interface {
	AddConnection()
	RemoveConnection()
	RecordRead(bytes int64)
	RecordWrite(bytes int64)
}

type ConnStats struct {
	connections    int64
	newConnections int64
	readLoops      int64
	writeLoops     int64
	bytesRead      int64
	bytesWritten   int64
	lock           *sync.Mutex
}

func NewConnStats() *ConnStats {
	return &ConnStats{
		lock: &sync.Mutex{},
	}
}

func (c *ConnStats) Connections() int64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.connections
}

func (c *ConnStats) NewConnections() int64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.newConnections
}

func (c *ConnStats) ReadLoops() int64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.readLoops
}

func (c *ConnStats) WriteLoops() int64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.writeLoops
}

func (c *ConnStats) BytesRead() int64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.bytesRead
}

func (c *ConnStats) BytesWritten() int64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.bytesWritten
}

func (c *ConnStats) Clear() {
	c.ResetCounters()
}

func (s *ConnStats) ResetCounters() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.newConnections = 0
	s.readLoops = 0
	s.writeLoops = 0
	s.bytesRead = 0
	s.bytesWritten = 0
}

func (s *ConnStats) AddConnection() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.connections++
	s.newConnections++
}

func (s *ConnStats) RemoveConnection() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.connections--
}

func (s *ConnStats) RecordRead(bytes int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.readLoops++
	s.bytesRead += bytes
}

func (s *ConnStats) RecordWrite(bytes int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.writeLoops++
	s.bytesWritten += bytes
}

func (s *ConnStats) Update(update func(stats *ConnStats)) {
	s.lock.Lock()
	defer s.lock.Unlock()
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
