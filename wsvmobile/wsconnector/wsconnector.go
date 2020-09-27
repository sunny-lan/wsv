package wsconnector

import (
	"github.com/gorilla/websocket"
	"github.com/sunny-lan/wsv/common"
	"sync"
	"time"
)

type lockedWs struct {
	lock *sync.Mutex
	ws   *websocket.Conn
}

//WsConnector handles connections from tun converting them to websocket
type WsConnector struct {
	server       string
	udpWs        *sync.Map
	writeTimeout time.Duration //TODO actually have timeout
	klist        common.KillList
}

// msg is a general purpose struct use to send control messages to the server
// TODO make this more efficient, go streamlined
type msg struct {
	ConType string `json:"ConType"`
	Dst     string `json:"Dst"`
	Port    int    `json:"Port"`
}

// NewWsConnector creates a new instance of WsConnector
// which connects to the websocket server given by server
// it assumes the server follows the wsv protocol
func NewWsConnector(server string) *WsConnector {
	return &WsConnector{
		server,
		&sync.Map{},
		time.Second * 5, //TODO actually use this.
		common.NewKillList(),
	}
}

// Close closes the WsConnector
// DOES NOT return internal errors during closing
func (t WsConnector) Close() {
	t.klist.KillAll()
}
