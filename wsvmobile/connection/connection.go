package connection

import "io"

type Connection interface {
	io.ReadWriteCloser
}

type Dialer interface {
	Dial() (Connection, error)
}
