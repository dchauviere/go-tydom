package api

import (
	"fmt"
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

type devicesDataEndpoint struct {
	ID    int              `json:"id"`
	Error int              `json:"error"`
	Data  []map[string]any `json:"data"`
}

type DevicesData struct {
	ID        int                   `json:"id"`
	Endpoints []devicesDataEndpoint `json:"endpoints"`
}

/*
[
	{
		"id":0,
		"endpoints":[
			{
				"id":0,
				"error":0,
				"data":[
					{
						"name":"position",
						"validity":"upToDate",
						"value":100
					}
				]
			}
		]
	}
]
*/

type KeyValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func parseDevicesData(resp *http.Response) (*[]DevicesData, error) {
	var data []DevicesData

	err := parse(resp, &data)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to parse Info response"}
	}

	return &data, nil
}

// Get all devices data.
func (ac *ClientAPI) GetDevicesData() (*[]DevicesData, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/devices/data", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send devices data request"}
	}

	return parseDevicesData(resp)
}

func (asc *ClientAsyncAPI) GetDevicesData(hook func(data *[]DevicesData)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/devices/data", "", func(resp *http.Response) {
		data, err := parseDevicesData(resp)
		if err != nil {
			asc.Client.Logger.Error("bad devices data response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request devices data"}
	}

	return nil
}

// Get data from a device.
func (ac *ClientAPI) GetDeviceData(deviceID string, endpointID string) (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", fmt.Sprintf("/devices/{%s}/endpoints/{%s}/data", deviceID, endpointID), "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send device data request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) GetDeviceData(deviceID string, endpointID string, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest(
		"GET",
		fmt.Sprintf("/devices/{%s}/endpoints/{%s}/data", deviceID, endpointID),
		"",
		func(resp *http.Response) {
			data, err := parseAny(resp)
			if err != nil {
				asc.Client.Logger.Error("bad device data response")

				return
			}

			hook(data)
		},
	)
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request device data"}
	}

	return nil
}

// Set device data.
func (ac *ClientAPI) SetDeviceData(deviceID string, endpointID string, kv KeyValue) (*any, error) {
	resp, err := ac.Client.HTTPRequest(
		"PUT",
		fmt.Sprintf("/devices/%s/endpoints/%s/data", deviceID, endpointID),
		fmt.Sprintf(`[{"name":"%s","value":"%s"}]`, kv.Name, kv.Value),
	)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send device data set : " + err.Error()}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) SetDeviceData(deviceID string, endpointID string, kv KeyValue, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest(
		"PUT",
		fmt.Sprintf("/devices/%s/endpoints/%s/data", deviceID, endpointID),
		fmt.Sprintf(`[{"name":"%s","value":"%s"}]`, kv.Name, kv.Value),
		func(resp *http.Response) {
			data, err := parseAny(resp)
			if err != nil {
				asc.Client.Logger.Error("bad set device data response")

				return
			}

			hook(data)
		},
	)
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request device data set"}
	}

	return nil
}
