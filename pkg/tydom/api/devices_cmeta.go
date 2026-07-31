package api

import (
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

type CMetadataParameters struct {
	Name       string `json:"name"`
	EnumValues []string
}

type CMetadata struct {
	Name       string              `json:"name"`
	Parameters CMetadataParameters `json:"parameters"`
}

type DevicesCMetaDataEndpoint struct {
	ID        int         `json:"id"`
	Error     int         `json:"error"`
	CMetadata []CMetadata `json:"cmetadata"`
}

type DevicesCMetaData struct {
	ID        int                        `json:"id"`
	Endpoints []DevicesCMetaDataEndpoint `json:"endpoints"`
}

/*
[
	{
		"id: "device_id",
		"endpoints": {
			"id": "endpoint_id",
			"cmetadata": {
				"name": "",
				"parameters": {
					"name": "",
					"enum_values": []
				}
			}
		}
	}
]
*/

func parseDevicesCMetaData(resp *http.Response) (*[]DevicesCMetaData, error) {
	var data []DevicesCMetaData

	err := parse(resp, &data)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to parse devices cmetadata response"}
	}

	return &data, nil
}

// Get metadata configuration to list poll devices (like Tywatt).
func (ac *ClientAPI) GetDeviceConfigMetadata() (*[]DevicesCMetaData, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/devices/cmeta", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send devices cmetadata request"}
	}

	return parseDevicesCMetaData(resp)
}

func (asc *ClientAsyncAPI) GetDeviceConfigMetadata(hook func(data *[]DevicesCMetaData)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/devices/cmeta", "", func(resp *http.Response) {
		data, err := parseDevicesCMetaData(resp)
		if err != nil {
			asc.Client.Logger.Error("bad devices cmetadata response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request devices cmetadata"}
	}

	return nil
}
