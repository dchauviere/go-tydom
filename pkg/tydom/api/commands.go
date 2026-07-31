package api

import (
	"fmt"
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

// Send Ping.
func (ac *ClientAPI) Ping() (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/ping", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send ping request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) Ping(hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/ping", "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad ping response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request ping"}
	}

	return nil
}

// Request gateway api mode.
func (ac *ClientAPI) SetGatewayAPIMode() error {
	resp, err := ac.Client.HTTPRequest("PUT", "/configs/gateway/api_mode", "")
	if err != nil {
		return &tydom.Error{Msg: "failed to send gateway api mode request"}
	}

	if resp.StatusCode != http.StatusOK {
		return &tydom.Error{Msg: fmt.Sprintf("set api mode command failed (http_code: %d)", resp.StatusCode)}
	}

	return nil
}

func (asc *ClientAsyncAPI) SetGatewayAPIMode(hook func(status int)) error {
	err := asc.Client.HTTPAsyncRequest("PUT", "/configs/gateway/api_mode", "", func(resp *http.Response) {
		hook(resp.StatusCode)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request gateway api mode"}
	}

	return nil
}

// Request refresh.
func (ac *ClientAPI) RefreshAll() error {
	resp, err := ac.Client.HTTPRequest("POST", "/refresh/all", "")
	if err != nil {
		return &tydom.Error{Msg: "failed to send refresh request"}
	}

	if resp.StatusCode != http.StatusOK {
		return &tydom.Error{Msg: fmt.Sprintf("refresh command failed (http_code: %d)", resp.StatusCode)}
	}

	return nil
}

func (asc *ClientAsyncAPI) RefreshAll(hook func(status int)) error {
	err := asc.Client.HTTPAsyncRequest("POST", "/refresh/all", "", func(resp *http.Response) {
		hook(resp.StatusCode)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request refresh"}
	}

	return nil
}

// Request Update Firmware.
func (ac *ClientAPI) UpdateFirmware() (*any, error) {
	resp, err := ac.Client.HTTPRequest("PUT", "/configs/gateway/update", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send firmware update"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) UpdateFirmware(hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("PUT", "/configs/gateway/update", "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad firmware update response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request update firmware"}
	}

	return nil
}
