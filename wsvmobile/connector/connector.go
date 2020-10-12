package connector

import (
	"github.com/sunny-lan/wsv/common"
	"github.com/sunny-lan/wsv/wsvmobile/connection"
	"sync"
)

type lockedConn struct {
	lock *sync.Mutex
	ws   *connection.Connection
}

//TODO use errorf for everything
//Connector handles connections from tun converting them to websocket
type Connector struct {
	Stats *wsConnStats

	udpWs *sync.Map
	kList common.KillList //TODO use once

	dialer connection.Dialer
}

// msg is a general purpose struct use to send control messages to the originalServer
// TODO make this more efficient, go streamlined
type msg struct {
	ConType string `json:"ConType"`
	Dst     string `json:"Dst"`
	Port    int    `json:"Port"`
}

// NewConnector creates a new instance of Connector
// which connects to the websocket originalServer given by originalServer
// it assumes the originalServer follows the wsv protocol
func NewConnector(server string, dialer connection.Dialer) (*Connector, error) {
	return &Connector{
		Stats:  newWsConnStats(),
		udpWs:  &sync.Map{},
		kList:  common.NewKillList(),
		dialer: dialer,
	}, nil
}

// Close closes the Connector
// DOES NOT return internal errors during closing
func (t *Connector) Close() {
	t.kList.KillAll()
}
