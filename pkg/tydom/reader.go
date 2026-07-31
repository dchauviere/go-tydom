package tydom

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func (tc *Client) notifyDisconnect() {
	tc.reconnect <- struct{}{}
}

func (tc *Client) readMessage() {
	if err := tc.websocket.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		tc.Logger.Error("error setting read deadline on websocket", "error", err)
		tc.notifyDisconnect()

		return
	}

	messageType, r, err := tc.websocket.NextReader()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			return
		}
		tc.Logger.Error("error reading message", "error", err)
		tc.notifyDisconnect()
		fmt.Println()
		return
	}

	message, err := io.ReadAll(r)
	if err != nil {
		tc.Logger.Error("failed to read message", "error", err)
		return
	}

	if messageType != websocket.BinaryMessage {
		tc.Logger.Debug("received non binary message", "message", message)
		return
	}

	tc.Logger.Debug("get a non empty message", "msg", string(message))

	rawData := bufio.NewReader(bytes.NewBuffer(message))

	resp, err := http.ReadResponse(rawData, nil)
	if err != nil { // this is a server request
		rawData = bufio.NewReader(bytes.NewBuffer(message))
		req, err := http.ReadRequest(rawData)
		if err != nil { // this is nothing usable
			tc.Logger.Error("parsing error", "error", err, "raw", message)

			return
		}

		tc.requestHookQueue <- req

		return
	}

	if channel, exists := tc.getSyncResponseChannel(resp.Header.Get("Transac-Id")); exists {
		channel <- resp

		return
	}
	tc.asyncResponseQueue <- resp
}
