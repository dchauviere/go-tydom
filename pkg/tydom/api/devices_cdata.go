package api

import (
	"fmt"
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

type devicesCDataEndpoint struct {
	ID    int              `json:"id"`
	Error int              `json:"error"`
	Data  []map[string]any `json:"data"`
}

type DevicesCData struct {
	ID        int                    `json:"id"`
	Endpoints []devicesCDataEndpoint `json:"endpoints"`
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

// Aknownledge alarm event PUT /devices/xxxx/endpoints/xxxx/cdata?name=ackEventCmd HTTP/1.1 {"pwd":"xxxxxx"}
// get historical events GET /devices/xxxx/endpoints/xxxx/cdata?name=histo&type=ALL&indexStart=0&nbElem=10

// Credits to @mgcrea on github !
// AWAY # "PUT /devices/{}/endpoints/{}/cdata?name=alarmCmd HTTP/1.1\r\ncontent-length: 29\r\ncontent-type: application/json; charset=utf-8\r\ntransac-id: request_124\r\n\r\n\r\n{"value":"ON","pwd":{}}\r\n\r\n"
// HOME "PUT /devices/{}/endpoints/{}/cdata?name=zoneCmd HTTP/1.1\r\ncontent-length: 41\r\ncontent-type: application/json; charset=utf-8\r\ntransac-id: request_46\r\n\r\n\r\n{"value":"ON","pwd":"{}","zones":[1]}\r\n\r\n"
// DISARM "PUT /devices/{}/endpoints/{}/cdata?name=alarmCmd HTTP/1.1\r\ncontent-length: 30\r\ncontent-type: application/json; charset=utf-8\r\ntransac-id: request_7\r\n\r\n\r\n{"value":"OFF","pwd":"{}"}\r\n\r\n"
// PUT /devices/{}/endpoints/{}/cdata?name=alarmCmd
//   HTTP/1.1\nContent-Length: 32\nContent-Type: application/json; charset=UTF-8\nTransac-Id: 1739979111409\nUser-Agent: Jakarta Commons-HttpClient/3.1\nHost: mediation.tydom.com:443
//   {"pwd":"######","value":"PANIC"}

// variables:
// id
// Cmd
// value
// pwd
// zones

func parseDevicesCData(resp *http.Response) (*[]DevicesCData, error) {
	var data []DevicesCData

	err := parse(resp, &data)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to parse Info response"}
	}

	return &data, nil
}

// Get all devices cdata.
func (ac *ClientAPI) GetDevicesCData() (*[]DevicesCData, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/devices/cdata", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send devices cdata request"}
	}

	return parseDevicesCData(resp)
}

func (asc *ClientAsyncAPI) GetDevicesCData(hook func(data *[]DevicesCData)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/devices/cdata", "", func(resp *http.Response) {
		data, err := parseDevicesCData(resp)
		if err != nil {
			asc.Client.Logger.Error("bad devices cdata response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request devices cdata"}
	}

	return nil
}

// Get cdata from a device.
func (ac *ClientAPI) GetDeviceCData(deviceID string, endpointID string) (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", fmt.Sprintf("/devices/{%s}/endpoints/{%s}/cdata", deviceID, endpointID), "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send device cdata request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) GetDeviceCData(deviceID string, endpointID string, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest(
		"GET",
		fmt.Sprintf("/devices/{%s}/endpoints/{%s}/cdata", deviceID, endpointID),
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

// Set device cdata.
func (ac *ClientAPI) SetDeviceCData(deviceID string, endpointID string, kv KeyValue) (*any, error) {
	resp, err := ac.Client.HTTPRequest(
		"PUT",
		fmt.Sprintf("/devices/{%s}/endpoints/{%s}/cdata", deviceID, endpointID),
		fmt.Sprintf(`[{"name":"%s","value":"%s"}]`, kv.Name, kv.Value),
	)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send device cdata set"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) SetDeviceCData(deviceID string, endpointID string, kv KeyValue, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest(
		"PUT",
		fmt.Sprintf("/devices/{%s}/endpoints/{%s}/cdata", deviceID, endpointID),
		fmt.Sprintf(`[{"name":"%s","value":"%s"}]`, kv.Name, kv.Value),
		func(resp *http.Response) {
			data, err := parseAny(resp)
			if err != nil {
				asc.Client.Logger.Error("bad set device cdata response")

				return
			}

			hook(data)
		},
	)
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request device cdata set"}
	}

	return nil
}
