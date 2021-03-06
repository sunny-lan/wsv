package wsvmobile

import (
	"errors"
	"fmt"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/eycorsican/go-tun2socks/common/log/simple"
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/imdario/mergo"
	"github.com/sunny-lan/wsv/common"
	"github.com/sunny-lan/wsv/wsvmobile/wsconnector"
	"io"
	"os"
	"syscall"
)

var lwipStack core.LWIPStack
var wsConn *wsconnector.WsConnector
var running = false
var wrappedTun io.ReadCloser

var (
	errInvalidFD      = errors.New("invalid FD")
	errAlreadyRunning = errors.New("already running")
	errNotRunning     = errors.New("not running")
)

type WsConnSettings wsconnector.Settings

type Settings struct {
	WsConnectionSettings *WsConnSettings
}

var defaultSettings = Settings{}

// Begin begins piping information from the given tun file descriptor
// to the given proxy host through ws
// It blocks until the one of the following happens:
// If requested to close through the Close method, returns nil
// Otherwise returns an error if irrecoverable (TUN failed)
func Begin(tunFD int64, proxyHost string, settings *Settings) error {
	log.RegisterLogger(simple.NewSimpleLogger())
	log.SetLevel(log.INFO)

	if running {
		return errAlreadyRunning
	}

	var cpy = defaultSettings
	e := mergo.Merge(&cpy, settings, mergo.WithOverride)
	if e != nil {
		panic(fmt.Errorf("merge wsv settings failed %w", e))
	}

	running = true
	defer func() { running = false }()

	log.Infof("Go code running")

	var tun = os.NewFile(uintptr(tunFD), "tunFD")
	if tun == nil {
		log.Errorf("invalid tunFD")
		return errInvalidFD
	}

	wsConn, e := wsconnector.NewWsConnector(proxyHost, (*wsconnector.Settings)(settings.WsConnectionSettings))
	if e != nil {
		return e
	}
	defer func() {
		wsConn.Close()
		wsConn = nil
	}()

	lwipStack = core.NewLWIPStack()
	defer func() {
		e := lwipStack.Close()
		if e != nil {
			log.Errorf("unable to close lwipStack %v", e)
		}
		lwipStack = nil
	}()

	core.RegisterTCPConnHandler(wsConn)
	core.RegisterUDPConnHandler(wsConn)
	core.RegisterOutputFn(tun.Write)

	wrappedTun = common.WrapReaderEof(tun)
	defer func() {
		wrappedTun = nil
	}()

	_, e = io.Copy(lwipStack, wrappedTun)

	if e != nil {
		log.Errorf("Go Unexpected exit with error: %v", e)
		return fmt.Errorf("enexpected error while reading from TUN %w. go killed", e)
	}

	log.Infof("Go Expected exit with error: %v", e)
	return nil
}

// Close closes the current connection to proxy server, closing the following:
// TCP stack, TUN file descriptor, existing ws connections
// it DOES NOT return errors which occur during closing the stack or tun
func Close() error {
	if !running {
		return errNotRunning
	}
	log.Infof("Go stack requested to stop")
	return wrappedTun.Close()
}

func CloseFD(fd int) error {
	return syscall.Close(fd)
}

func Stats() wsconnector.WsConnStatsReader {
	return wsConn.Stats
}
