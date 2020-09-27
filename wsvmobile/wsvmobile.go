package wsvmobile

import (
	"errors"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/eycorsican/go-tun2socks/common/log/simple"
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/sunny-lan/wsv/common"
	"github.com/sunny-lan/wsv/wsvmobile/wsconnector"
	"io"
	"os"
	"syscall"
)

var lwipStack core.LWIPStack
var wsConn *wsconnector.WsConnector
var klist common.KillList
var running = false

var (
	ErrInvalidFD      = errors.New("invalid FD")
	ErrAlreadyRunning = errors.New("already running")
	ErrNotRunning     = errors.New("not running")
)

// Begin begins piping information from the given tun file descriptor
// to the given proxy host through ws
// It blocks until the one of the following happens:
// If requested to close through the Close method, returns nil
// Otherwise returns an error if irrecoverable (TUN failed)
func Begin(tunFD int64, proxyHost string) error {
	log.RegisterLogger(simple.NewSimpleLogger())
	log.SetLevel(log.INFO)

	if running {
		return ErrAlreadyRunning
	}

	running = true
	log.Infof("Go code running")

	klist = common.NewKillList()
	defer klist.KillAll()

	var tun = os.NewFile(uintptr(tunFD), "tunFD")
	if tun == nil {
		log.Errorf("invalid tunFD")
		return ErrInvalidFD
	}
	klist.AddKiller(tun, func() {
		log.Infof("Go: Killing tun")
		e := tun.Close()
		if e != nil {
			log.Errorf("unable to close tun %v", e)
		}
		e = syscall.Close(int(tunFD))
		if e != nil {
			log.Errorf("unable to close tun through syscall %v", e)
		}
	})

	wsConn = wsconnector.NewWsConnector(proxyHost)
	klist.AddKiller(wsConn, func() {
		wsConn.Close()
		wsConn = nil
	})

	lwipStack = core.NewLWIPStack()
	klist.AddKiller(lwipStack, func() {
		e := lwipStack.Close()
		if e != nil {
			log.Errorf("unable to close lwipStack %v", e)
		}
		lwipStack = nil
	})

	core.RegisterTCPConnHandler(wsConn)
	core.RegisterUDPConnHandler(wsConn)
	core.RegisterOutputFn(tun.Write)

	_, e := io.Copy(lwipStack, tun)

	if running {
		log.Errorf("Go Unexpected exit %v", e)
		return e
	} else {
		log.Infof("Go Expected exit, %v", e)
		return nil
	}
}

// Close closes the current connection to proxy server, closing the following:
// TCP stack, TUN file descriptor, existing ws connections
// it DOES NOT return errors which occur during closing the stack or tun
func Close() error {
	if !running {
		return ErrNotRunning
	}
	log.Infof("Go stack requested to stop")
	running = false
	klist.KillAll()
	return nil
}

func CloseFD(fd int) error {
	return syscall.Close(fd)
}
