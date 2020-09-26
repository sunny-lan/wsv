package wsvmobile

import (
	"errors"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/eycorsican/go-tun2socks/common/log/simple"
	"github.com/eycorsican/go-tun2socks/core"
	"io"
	"os"
	"syscall"
)

var lwipStack core.LWIPStack
var running = false

var (
	errInvalidFD      = errors.New("invalid FD")
	errAlreadyRunning = errors.New("already running")
	errNotRunning     = errors.New("not running")
)

// Begin begins piping information from the given tun file descriptor
// to the given proxy host through ws
func Begin(tunFD int64, proxyHost string) error {
	log.RegisterLogger(simple.NewSimpleLogger())
	log.SetLevel(log.INFO)
	log.Infof("Go code running")

	if running {
		return errAlreadyRunning
	}
	running = true

	var f = os.NewFile(uintptr(tunFD), "tunFD")

	if f == nil {
		log.Errorf("invalid tunFD")
		return errInvalidFD
	}
	var s = newTcpUdpHandler(proxyHost)
	lwipStack = core.NewLWIPStack()
	core.RegisterTCPConnHandler(s)
	core.RegisterUDPConnHandler(s)
	core.RegisterOutputFn(f.Write)

	_, e := io.Copy(lwipStack, f)
	log.Infof("Go code exit %v", e)

	return e
}

// Close closes the current connection to proxy server, killing and cleaning everything
func Close() error {
	if !running {
		return errNotRunning
	}
	e := lwipStack.Close()
	lwipStack = nil
	return e
}

// CloseFD closes the specified file descriptor
func CloseFD(tunFD int) error {
	return syscall.Close(tunFD)
}
