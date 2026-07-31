package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

type ClientAPI struct {
	Client *tydom.Client
}

type ClientAsyncAPI struct {
	Client *tydom.Client
}

type EmptyResponse any

func parse(resp *http.Response, data any) error {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return &tydom.Error{Msg: "fail to read response body", Err: err}
	}

	if string(body) != "" {
		err = json.Unmarshal(body, data)
		if err != nil {
			return &tydom.Error{Msg: fmt.Sprintf("failed to decode body (%s) (http_code: %d) (content-length: %s)", body, resp.StatusCode, resp.Header.Get("Content-Length")), Err: err}
		}
	}

	return nil
}

func parseAny(resp *http.Response) (*any, error) {
	var data any

	err := parse(resp, &data)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to parse response", Err: err}
	}

	return &data, nil
}

func (ac *ClientAPI) Get(path string) (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", path, "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send get request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) Get(path string, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("GET", path, "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad get response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request get"}
	}

	return nil
}

func (ac *ClientAPI) Post(path string, body string) (*any, error) {
	resp, err := ac.Client.HTTPRequest("POST", path, body)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send post request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) Post(path string, body string, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("POST", path, body, func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad post response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request post"}
	}

	return nil
}

func (ac *ClientAPI) Put(path string, body string) (*any, error) {
	resp, err := ac.Client.HTTPRequest("PUT", path, body)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send put request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) Put(path string, body string, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("PUT", path, body, func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad put response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request put"}
	}

	return nil
}

func (ac *ClientAPI) Delete(path string) (*any, error) {
	resp, err := ac.Client.HTTPRequest("DELETE", path, "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send delete request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) Delete(path string, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("DELETE", path, "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad delete response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request delete"}
	}

	return nil
}
