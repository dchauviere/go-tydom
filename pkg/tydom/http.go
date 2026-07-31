package tydom

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/dchauviere/go-tydom/internal/config"
	"github.com/google/uuid"
)

const (
	RequestTimeout int = 5
)

func makeRequest(method string, url string, body string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, &Error{Msg: "bad http method : " + err.Error()}
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Transac-Id", uuid.New().String())
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", config.USERAGENT, config.VERSION))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	return req, err
}

func (tc *Client) HTTPRequest(method string, path string, body string) (*http.Response, error) {
	tc.Logger.Debug("http request", "method", method, "path", path, "body", body)

	req, err := makeRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	responseChannel := make(chan *http.Response)
	tc.registerSyncResponse(req.Header.Get("Transac-Id"), responseChannel)
	tc.DataSender <- req

	for {
		select {
		case response := <-responseChannel:
			close(responseChannel)
			tc.unregisterSyncResponse(req.Header.Get("Transac-Id"))

			return response, nil
		case <-time.After(time.Duration(RequestTimeout) * time.Second):
			return nil, &Error{Msg: "timeout waiting for response"}
		}
	}
}

func (tc *Client) HTTPAsyncRequest(method string, path string, body string, hook func(resp *http.Response)) error {
	req, err := makeRequest(method, path, body)
	if err != nil {
		return err
	}

	tc.asyncResponseRegistry[req.Header.Get("Transac-Id")] = hook

	tc.DataSender <- req

	return nil
}
