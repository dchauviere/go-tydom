package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

// Get an area.
func (ac *ClientAPI) GetArea(areaID string) (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", fmt.Sprintf("/areas/%s/data", areaID), "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send areas request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) GetArea(areaID string, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("GET", fmt.Sprintf("/areas/%s/data", areaID), "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad area response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request area"}
	}

	return nil
}

// Set area.
func (ac *ClientAPI) SetArea(areaID string, kvList []KeyValue) (*any, error) {
	var body []byte

	var err error

	if body, err = json.Marshal(kvList); err != nil {
		return nil, &tydom.Error{Msg: "cannot encode kv list", Err: err}
	}

	resp, err := ac.Client.HTTPRequest("PUT", fmt.Sprintf("/areas/%s/data", areaID), string(body))
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send set area request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) SetArea(areaID string, kvList []KeyValue, hook func(data *any)) error {
	var body []byte

	var err error

	if body, err = json.Marshal(kvList); err != nil {
		return &tydom.Error{Msg: "cannot encode kv list", Err: err}
	}

	err = asc.Client.HTTPAsyncRequest(
		"PUT",
		fmt.Sprintf("/areas/%s/data", areaID),
		string(body),
		func(resp *http.Response) {
			data, err := parseAny(resp)
			if err != nil {
				asc.Client.Logger.Error("bad set area response")

				return
			}

			hook(data)
		},
	)
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request set area"}
	}

	return nil
}

// Request all areas.
func (ac *ClientAPI) GetAreas() (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/areas/data", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send areas request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) GetAreas(hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/areas/data", "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad areas response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request areas"}
	}

	return nil
}
