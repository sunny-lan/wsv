package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
)

type Connection struct {
	ws        *websocket.Conn
	curReader io.Reader
	readType  int
	writeType int
}

func newConnection(c *websocket.Conn) *Connection {
	return &Connection{
		ws:        c,
		curReader: nil,
		readType:  -1,
		writeType: -1,
	}
}

func (w *Connection) Expect(messageType int) {
	w.readType = messageType
}

func (w *Connection) SetMessageType(t int) {
	w.writeType = t
}

func (w *Connection) Read(p []byte) (n int, err error) {
	if w.curReader == nil {
		t, r, e := w.ws.NextReader()
		if e != nil {
			return 0, fmt.Errorf("failed to get ws next reader %w", e)
		}
		if w.readType != -1 && t != w.readType {
			return 0, fmt.Errorf("unexpected message type from ws")
		}
		w.curReader = r
	}
	n, e := w.curReader.Read(p)
	if e == io.EOF {
		return n, nil
	} else {
		return n, e
	}
}

func (w *Connection) Write(p []byte) (n int, err error) {
	if w.writeType == -1 {
		return 0, fmt.Errorf("must select message type before writing")
	}
	e := w.ws.WriteMessage(w.writeType, p)
	if e != nil {
		return 0, e
	} else {
		return len(p), nil
	}
}

func (w *Connection) Close() error {
	return w.Close()
}
