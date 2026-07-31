package tydom

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 10 * time.Second

	// http request timeout.
	httpRequestTimeout = 10 * time.Second

	// max backoff time for connection retry.
	maxBackoff = 60 * time.Second
)

type State int

const (
	Disconnected State = iota
	Disconnecting
	Connecting
	Connected
)

func (s State) String() string {
	switch s {
	case Disconnected:
		return "Disconnected"
	case Disconnecting:
		return "Disconnecting"
	case Connecting:
		return "Connecting"
	case Connected:
		return "Connected"
	default:
		return "Unknown"
	}
}

type Client struct {
	TargetHost                url.URL
	DataSender                chan *http.Request
	GatewayID                 string
	websocket                 *websocket.Conn
	Logger                    *slog.Logger
	syncResponseRegistry      map[string]chan *http.Response
	syncResponseRegistryMutex sync.Mutex
	asyncResponseRegistry     map[string]func(resp *http.Response)
	asyncResponseQueue        chan *http.Response
	requestHookRegistry       []RequestHook
	requestHookQueue          chan *http.Request
	shutdown                  chan struct{}
	reconnect                 chan struct{}
	wg                        sync.WaitGroup
	mutex                     sync.RWMutex
	state                     State
	password                  string
}

func (tc *Client) AddRequestHook(method string, path string, hook func([]byte)) {
	tc.requestHookRegistry = append(tc.requestHookRegistry, RequestHook{Method: method, Path: path, Hook: hook})
}

func (tc *Client) RemoveRequestHook(method string, path string) {
	for index, entry := range tc.requestHookRegistry {
		if entry.Method == strings.ToLower(method) && entry.Path == strings.ToLower(path) {
			tc.requestHookRegistry = append(tc.requestHookRegistry[:index], tc.requestHookRegistry[index+1:]...)

			return
		}
	}
}

func (tc *Client) registerSyncResponse(ID string, channel chan *http.Response) {
	tc.syncResponseRegistryMutex.Lock()
	defer tc.syncResponseRegistryMutex.Unlock()
	tc.syncResponseRegistry[ID] = channel
}

func (tc *Client) unregisterSyncResponse(ID string) {
	tc.syncResponseRegistryMutex.Lock()
	defer tc.syncResponseRegistryMutex.Unlock()
	delete(tc.syncResponseRegistry, ID)
}

func (tc *Client) getSyncResponseChannel(ID string) (chan *http.Response, bool) {
	tc.syncResponseRegistryMutex.Lock()
	defer tc.syncResponseRegistryMutex.Unlock()
	channel, ok := tc.syncResponseRegistry[ID]
	return channel, ok
}

func (tc *Client) setState(state State) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.state = state
}

func (tc *Client) State() State {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	return tc.state
}

func (tc *Client) IsConnected() bool {
	return tc.State() == Connected
}

func (tc *Client) connect() {
	backoff := time.Second
	websocket.DefaultDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402

	for {
		select {
		case <-tc.shutdown:
			return
		default:
			var header http.Header
			var err error
			if tc.password != "" {
				cred, err := tc.Authenticate(tc.password)
				if err != nil {
					tc.Logger.Error("error authenticating", "error", err.Error())
					time.Sleep(backoff)
					if backoff*2 >= maxBackoff {
						backoff = maxBackoff
					} else {
						backoff *= 2
					}

					continue
				}
				header = http.Header{
					"Authorization": []string{cred},
				}
			}

			var resp *http.Response
			tc.websocket, resp, err = websocket.DefaultDialer.Dial(tc.TargetHost.String(), header)

			if err != nil {
				tc.Logger.Error("error connecting websocket", "error", err.Error())
				tc.Logger.Debug("ws response", "data", resp)
				time.Sleep(backoff)
				if backoff*2 >= maxBackoff {
					backoff = maxBackoff
				} else {
					backoff *= 2
				}

				continue
			}
		}

		break
	}

	tc.websocket.SetPongHandler(func(string) error {
		tc.Logger.Debug("pong received")

		return nil
	})

	tc.Logger.Info("Tydom connected")
}

func (tc *Client) disconnect() {
	if tc.websocket == nil {
		tc.Logger.Debug("nothing to disconnect")

		return
	}

	tc.Logger.Debug("set write deadline")
	err := tc.websocket.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		tc.Logger.Error("fail to set write deadline on websocket", "error", err)
	}

	tc.Logger.Debug("sending close message")
	err = tc.websocket.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		tc.Logger.Error("failed to send close to websocket", "error", err)
	}

	tc.Logger.Debug("waiting close message ack")
	// Read messages until the close message is confirmed
	for {
		err = tc.websocket.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			tc.Logger.Error("failed to set read deadline on websocket", "error", err)
		}

		_, _, err = tc.websocket.NextReader()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			break
		}

		if err != nil {
			break
		}
	}

	tc.Logger.Debug("closing websocket")
	// Close the TCP connection
	err = tc.websocket.Close()
	if err != nil {
		tc.Logger.Error("failed to close websocket")
	}

	tc.Logger.Info("Tydom disconnected")
}

func (tc *Client) connectionMonitor() {
	defer tc.wg.Done()
	for {
		select {
		// Process shutdown signal
		case <-tc.shutdown:
			tc.Logger.Debug("shutting down connectionMonitor")
			tc.setState(Disconnecting)
			tc.disconnect()
			tc.setState(Disconnected)
			tc.Logger.Debug("connectionMonitor shutdown")

			return
		case <-tc.reconnect:
			tc.setState(Disconnecting)
			tc.disconnect()
			tc.setState(Connecting)
			tc.connect()
			tc.setState(Connected)
		}
	}
}

func (tc *Client) Start() error {
	if tc.State() != Disconnected {
		return &Error{Msg: "client already started"}
	}

	tc.Logger.Debug("start tydom client")

	tc.shutdown = make(chan struct{})
	tc.DataSender = make(chan *http.Request, 20)

	tc.wg.Add(5)
	go tc.processServerRequest()
	go tc.processAsyncResponse()
	go tc.runSender()
	go tc.runReceiver()
	go tc.connectionMonitor()

	tc.reconnect <- struct{}{}

	return nil
}

func (tc *Client) Stop() {
	close(tc.shutdown)

	tc.wg.Wait()

	tc.Logger.Debug("tydom client goroutines stopped")
}

func (tc *Client) sendPing() {
	tc.Logger.Debug("send ping")

	err := tc.websocket.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		tc.Logger.Error("fail to set write deadline on websocket", "error", err)
	}

	if err := tc.websocket.WriteMessage(websocket.PingMessage, nil); err != nil {
		tc.Logger.Error("failed to write ping message", "error", err)
	}
}

func (tc *Client) sendData(req *http.Request) {
	var buf bytes.Buffer

	if err := req.Write(&buf); err != nil {
		tc.Logger.Error("cannot serialize HTTP request", "error", err)
		return
	}

	tc.Logger.Info("sending data", "data", buf.String())

	err := tc.websocket.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		tc.Logger.Error("failed to set write deadline for tydom sending", "error", err)
		return
	}

	writer, err := tc.websocket.NextWriter(websocket.BinaryMessage)
	if err != nil {
		tc.Logger.Error("failed to create next writer for tydom sending", "error", err)
		return
	}

	if _, err := writer.Write(buf.Bytes()); err != nil {
		tc.Logger.Error("failed to write for tydom sending", "error", err)
		return
	}

	if err := writer.Close(); err != nil {
		tc.Logger.Error("websocket error", "error", err)
	}
}

func (tc *Client) runSender() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		tc.wg.Done()
	}()

	for {
		select {
		case <-tc.shutdown:
			tc.Logger.Debug("runSender shutdown")

			return
		case req, ok := <-tc.DataSender:
			if ok && tc.State() == Connected {
				tc.sendData(req)
			}
		// WebSocket Heartbeat
		case <-ticker.C:
			if tc.State() == Connected {
				tc.sendPing()
			}
		}
	}
}

func (tc *Client) runReceiver() {
	defer func() {
		for key, value := range tc.syncResponseRegistry {
			close(value)
			delete(tc.syncResponseRegistry, key)
		}
		tc.wg.Done()
	}()

	for {
		select {
		case <-tc.shutdown:
			tc.Logger.Debug("runReceiver shutdown")

			return
		default:
			if tc.State() == Connected {
				tc.readMessage()
			}
		}
	}
}

func NewClient(target string, gatewayID string, password string) *Client {
	targetURL := url.URL{
		Scheme:   "wss",
		Host:     target,
		Path:     "/mediation/client",
		RawQuery: fmt.Sprintf("mac=%s&appli=1", gatewayID),
	}

	return &Client{
		TargetHost:            targetURL,
		GatewayID:             gatewayID,
		Logger:                slog.With(slog.String("module", "TydomClient")),
		syncResponseRegistry:  make(map[string]chan *http.Response),
		asyncResponseRegistry: make(map[string]func(resp *http.Response)),
		asyncResponseQueue:    make(chan *http.Response, 10),
		requestHookQueue:      make(chan *http.Request, 10),
		reconnect:             make(chan struct{}),
		password:              password,
		state:                 Disconnected,
	}
}
