package api

import (
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

//nolint:tagliatelle
type MetaData struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Permission string   `json:"permission"`
	Validity   string   `json:"validity"`
	EnumValues []string `json:"enum_values"`
	Min        int      `json:"min"`
	Max        int      `json:"max"`
	Step       int      `json:"step"`
	Unit       string   `json:"unit"`
}

type EndpointMeta struct {
	ID       int        `json:"id"`
	Error    int        `json:"error"`
	Metadata []MetaData `json:"metadata"`
}

type DevicesMeta struct {
	ID        int `json:"id"`
	Endpoints []EndpointMeta
}

/*
[
	{
		"id":0,
		"endpoints":[
			{
				"id":0,
				"error":15,
				"metadata":[
					{
						"name":"positionCmd",
						"type":"string",
						"permission":"w",
						"validity":"INFINITE",
						"enum_values":[
							"DOWN",
							"UP",
							"STOP",
							"FAVORIT1",
							"FAVORIT2",
							"UP_SLOW",
							"DOWN_SLOW"
						]
					},
					{
						"name":"thermicDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"STATUS_POLLING",
						"unit":"boolean"
					},
					{
						"name":"position",
						"type":"numeric",
						"permission":"rw",
						"validity":"DATA_POLLING",
						"min":0,
						"max":100,
						"step":1,
						"unit":"%"
					},
					{
						"name":"recFav",
						"type":"string",
						"permission":"w",
						"validity":"INFINITE",
						"enum_values":[
							"FAVORIT1",
							"FAVORIT2"
						]
					},
					{
						"name":"onFavPos",
						"type":"boolean",
						"permission":"r",
						"validity":"STATUS_POLLING",
						"unit":"boolean"
					},
					{
						"name":"upDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"STATUS_POLLING",
						"unit":"boolean"
					},
					{
						"name":"downDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"STATUS_POLLING",
						"unit":"boolean"
					},
					{
						"name":"obstacleDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"STATUS_POLLING",
						"unit":"boolean"
					},
					{
						"name":"intrusion",
						"type":"boolean",
						"permission":"r",
						"validity":"STATUS_POLLING",
						"unit":"boolean"
					},
					{
						"name":"battDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"STATUS_POLLING",
						"unit":"boolean"
					},
					{
						"name":"localisation",
						"type":"string",
						"permission":"w",
						"validity":"INFINITE",
						"enum_values":[
							"START"
						]
					},
					{
						"name":"modeAsso",
						"type":"string",
						"permission":"w",
						"validity":"INFINITE",
						"enum_values":[
							"START"
						]
					}
				]
			}
		]
	},
	{
		"id":3,
		"endpoints":[
			{
				"id":0,
				"error":0,
				"metadata":[
					{
						"name":"positionCmd",
						"type":"string",
						"permission":"w",
						"validity":"INFINITE",
						"enum_values":[
							"DOWN",
							"UP",
							"STOP",
							"FAVORIT1",
							"FAVORIT2",
							"UP_SLOW",
							"DOWN_SLOW"
						]
					},
					{
						"name":"thermicDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"ES_SUPERVISION",
						"unit":"boolean"
					},
					{
						"name":"position",
						"type":"numeric",
						"permission":"rw",
						"validity":"ES_SUPERVISION",
						"min":0,
						"max":100,
						"step":1,
						"unit":"%"
					},
					{
						"name":"recFav",
						"type":"string",
						"permission":"w",
						"validity":"INFINITE",
						"enum_values":[
							"FAVORIT1",
							"FAVORIT2"
						]
					},
					{
						"name":"onFavPos",
						"type":"boolean",
						"permission":"r",
						"validity":"ES_SUPERVISION",
						"unit":"boolean"
					},
					{
						"name":"upDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"ES_SUPERVISION",
						"unit":"boolean"
					},
					{
						"name":"downDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"ES_SUPERVISION",
						"unit":"boolean"
					},
					{
						"name":"obstacleDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"ES_SUPERVISION",
						"unit":"boolean"
					},
					{
						"name":"intrusion",
						"type":"boolean",
						"permission":"r",
						"validity":"ES_SUPERVISION",
						"unit":"boolean"
					},
					{
						"name":"battDefect",
						"type":"boolean",
						"permission":"r",
						"validity":"ES_SUPERVISION",
						"unit":"boolean"
					},
					{
						"name":"localisation",
						"type":"string",
						"permission":"w",
						"validity":"INFINITE",
						"enum_values":["START"]
					},
					{
						"name":"modeAsso",
						"type":"string",
						"permission":"w",
						"validity":"INFINITE",
						"enum_values":["START"]
					}
				]
			}
		]
	}
]
*/

func parseDevicesMeta(resp *http.Response) (*[]DevicesMeta, error) {
	var data []DevicesMeta

	err := parse(resp, &data)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to parse Info response"}
	}

	return &data, nil
}

// Get all devices metadata.
func (ac *ClientAPI) GetDeviceMetadata() (*[]DevicesMeta, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/devices/meta", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send devices metadata request"}
	}

	return parseDevicesMeta(resp)
}

func (asc *ClientAsyncAPI) GetDeviceMetadata(hook func(data *[]DevicesMeta)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/devices/meta", "", func(resp *http.Response) {
		data, err := parseDevicesMeta(resp)
		if err != nil {
			asc.Client.Logger.Error("bad devices metadata response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request devices metadata"}
	}

	return nil
}
