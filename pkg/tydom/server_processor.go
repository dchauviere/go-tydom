package tydom

import (
	"fmt"
	"io"
	"net/http"
)

type RequestHook struct {
	Method string
	Path   string
	Hook   func([]byte)
}

func (tc *Client) processRequest(req *http.Request) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return &Error{Msg: fmt.Sprintf("failed to read request body (%v)", err)}
	}

	defer req.Body.Close()
	tc.Logger.Debug("processRequest", "message", body)

	for _, hook := range tc.requestHookRegistry {
		if hook.Method == req.Method && hook.Path == req.URL.Path {
			hook.Hook(body)

			return nil
		}
	}

	tc.Logger.Debug("no hook for server request", "method", req.Method, "path", req.URL.Path, "data", body)

	return nil
}

func (tc *Client) processServerRequest() {
	defer tc.wg.Done()

	for {
		select {
		// Process shutdown signal
		case <-tc.shutdown:
			tc.Logger.Debug("processServerRequest shutdown")

			return
		case req := <-tc.requestHookQueue:
			tc.Logger.Debug("processing server request", "request", req)

			if err := tc.processRequest(req); err != nil {
				tc.Logger.Error("failed processing server request")
			}
		}
	}
}
