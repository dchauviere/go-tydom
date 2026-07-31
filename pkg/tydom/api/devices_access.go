package api

import (
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

type Addr struct {
	Net  int `json:"net"`
	Slot int `json:"slot"`
	ES   int `json:"es"`
}

type AccessInfo struct {
	Protocol string `json:"protocol"`
	Profile  string `json:"profile"`
	Type     string `json:"type"`
	Addr     Addr   `json:"addr"`
	SubAddr  int    `json:"subAddr"`
}

type AccessEndpoint struct {
	ID     int        `json:"id"`
	Error  int        `json:"error"`
	Access AccessInfo `json:"access"`
}

type DeviceAccess struct {
	ID        int              `json:"id"`
	Endpoints []AccessEndpoint `json:"endpoints"`
}

func parseDevicesAccess(resp *http.Response) (*[]DeviceAccess, error) {
	var data []DeviceAccess

	err := parse(resp, &data)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to parse device access response"}
	}

	return &data, nil
}

// Get Device access.
func (ac *ClientAPI) GetDevicesAccess() (*[]DeviceAccess, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/devices/access", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send devices access request"}
	}

	return parseDevicesAccess(resp)
}

func (asc *ClientAsyncAPI) GetDevicesAccess(hook func(data *[]DeviceAccess)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/devices/access", "", func(resp *http.Response) {
		data, err := parseDevicesAccess(resp)
		if err != nil {
			asc.Client.Logger.Error("bad devices access response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request info"}
	}

	return nil
}
