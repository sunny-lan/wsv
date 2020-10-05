package wsconnector

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/gorilla/websocket"
	"github.com/sunny-lan/wsv/common"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type lockedWs struct {
	lock *sync.Mutex
	ws   *websocket.Conn
}

type settings struct {
	connectTimeout time.Duration //TODO actually have timeout
	bufferSize     int64
	trustedCerts   []byte
}

//TODO use errorf for everything
//TODO split into udp vs tcp
//WsConnector handles connections from tun converting them to websocket
type WsConnector struct {
	Stats *wsConnStats

	udpWs    *sync.Map
	settings settings
	kList    common.KillList //TODO use once

	dialLock       *sync.Mutex
	concurrentMode *atomic.Value
	originalServer *url.URL
	workingServer  *atomic.Value
	dialer         *websocket.Dialer
}

// msg is a general purpose struct use to send control messages to the originalServer
// TODO make this more efficient, go streamlined
type msg struct {
	ConType string `json:"ConType"`
	Dst     string `json:"Dst"`
	Port    int    `json:"Port"`
}

func parseURL(s string) (*url.URL, error) {
	newURL, e := url.Parse(s)
	if e != nil {
		return nil, e
	}

	if newURL.Scheme == "http" {
		log.Warnf("url is http, converting to ws: %v", newURL)
		newURL.Scheme = "ws"
	}

	if newURL.Scheme == "https" {
		log.Warnf("url is https, converting to wss: %v", newURL)
		newURL.Scheme = "wss"
	}

	if !(newURL.Scheme == "wss" || newURL.Scheme == "ws") {
		return nil, fmt.Errorf("non-ws url: %v", newURL)
	}

	return newURL, nil
}

func (t *WsConnector) dialWs() (*websocket.Conn, *http.Response, error) {
	if t.concurrentMode.Load() == false {
		t.dialLock.Lock()
		defer t.dialLock.Unlock()
	}

	ws, resp, e := t.dialer.Dial(t.workingServer.Load().(*url.URL).String(), nil)
	if e == nil {
		//connection success, allow concurrent connections
		t.concurrentMode.Store(true)
		return ws, resp, nil
	}

	// dial failed, reset to original server, ban concurrent mode
	t.concurrentMode.Store(false)
	t.workingServer.Store(t.originalServer)

	// if ErrBadHandshake, we might have gotten a redirect
	if e == websocket.ErrBadHandshake {
		if resp == nil {
			panic("this should never happen. bug in websockets library")
		}
		if resp.StatusCode == 302 {
			location := resp.Header.Get("Location")
			log.Infof("try redirect to %v", location)

			newURL, e := parseURL(location)
			if e != nil {
				return nil, nil, fmt.Errorf("failed to parse redirect dest %v %w", location, e)
			}

			ws, resp, e = t.dialer.Dial(newURL.String(), nil)
			if e != nil {
				return nil, nil, fmt.Errorf("failed to dial redirected url: %v %w", t.workingServer, e)
			}

			//connection success, allow concurrent connections
			log.Infof("redirect success, storing as new location")
			t.concurrentMode.Store(true)
			t.workingServer.Store(newURL)
			return ws, resp, nil
		} else {
			return nil, nil, fmt.Errorf("unexpected status code upon dialing (expected redirect): %v", resp)
		}
	} else {
		return nil, nil, e
	}
}

func trustCert(cert []byte) *tls.Config {
	insecure := flag.Bool("insecure-ssl", false, "Accept/Ignore all originalServer SSL certificates")
	flag.Parse()

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
		log.Infof("No certs appended, using system certs only")
	}

	// Trust the augmented cert pool in our client
	return &tls.Config{
		InsecureSkipVerify: *insecure,
		RootCAs:            rootCAs,
	}

}

// NewWsConnector creates a new instance of WsConnector
// which connects to the websocket originalServer given by originalServer
// it assumes the originalServer follows the wsv protocol
func NewWsConnector(server string) (*WsConnector, error) {
	u, e := parseURL(server)
	if e != nil {
		return nil, e
	}
	settings := settings{
		connectTimeout: time.Second * 5,
		bufferSize:     32 * 1024,
	}
	dialer := websocket.DefaultDialer
	dialer.WriteBufferSize = int(settings.bufferSize)
	dialer.ReadBufferSize = int(settings.bufferSize)
	if settings.trustedCerts != nil {
		dialer.TLSClientConfig = trustCert(settings.trustedCerts)
	}
	c := &atomic.Value{}
	c.Store(false)
	w := &atomic.Value{}
	w.Store(u)
	return &WsConnector{
		Stats:          newWsConnStats(),
		udpWs:          &sync.Map{},
		settings:       settings,
		kList:          common.NewKillList(),
		dialLock:       &sync.Mutex{},
		concurrentMode: c,
		originalServer: u,
		workingServer:  w,
		dialer:         dialer,
	}, nil
}

// Close closes the WsConnector
// DOES NOT return internal errors during closing
func (t WsConnector) Close() {
	t.kList.KillAll()
}
