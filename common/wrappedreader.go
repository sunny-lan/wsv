package common

import (
	"errors"
	"io"
	"sync/atomic"
)

func WrapReaderEof(reader io.Reader) io.ReadCloser {
	x := &wrappedReader{
		atomic.Value{},
		reader,
	}
	x.closed.Store(false)
	return x
}

type wrappedReader struct {
	closed        atomic.Value
	wrappedReader io.Reader
}

func (r *wrappedReader) Close() error {
	if r.closed.Load() == true {
		return errors.New("reader already closed")
	}
	r.closed.Store(true)
	return nil
}

func (r *wrappedReader) Read(buf []byte) (int, error) {
	if r.closed.Load() == true {
		return 0, io.EOF
	} else {
		return r.wrappedReader.Read(buf)
	}
}
