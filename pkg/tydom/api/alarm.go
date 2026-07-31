package api

import (
	"fmt"
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

// Request refresh.
func (ac *ClientAPI) SetAlarmCData(deviceID, alarmID, value, zoneID, alarmPin string) (*any, error) {
	var body, cmd string

	switch {
	case value == "ACK":
		cmd = "ackEventCmd"
		body = fmt.Sprintf("{\"pwd\":\"%s\"}", alarmPin)
	case zoneID == "":
		cmd = "alarmCmd"
		body = fmt.Sprintf("{\"value\":\"%s\",\"pwd\":\"%s\"}", value, alarmPin)
	default:
		cmd = "zoneCmd"
		body = fmt.Sprintf("{\"value\":\"%s\",\"pwd\":\"%s\",\"zones\":[\"%s\"]}", value, alarmPin, zoneID)
	}

	resp, err := ac.Client.HTTPRequest(
		"PUT",
		fmt.Sprintf("/devices/%s/endpoints/%s/cdata?name=%s", deviceID, alarmID, cmd),
		body,
	)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send alarm cdata request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) SetAlarmCData(deviceID, alarmID, value, zoneID, pin string, hook func(data *any)) error {
	var body, cmd string

	switch {
	case value == "ACK":
		cmd = "ackEventCmd"
		body = fmt.Sprintf("{\"pwd\":\"%s\"}", pin)
	case zoneID == "":
		cmd = "alarmCmd"
		body = fmt.Sprintf("{\"value\":\"%s\",\"pwd\":\"%s\"}", value, pin)
	default:
		cmd = "zoneCmd"
		body = fmt.Sprintf("{\"value\":\"%s\",\"pwd\":\"%s\",\"zones\":[\"%s\"]}", value, pin, zoneID)
	}

	err := asc.Client.HTTPAsyncRequest(
		"PUT",
		fmt.Sprintf("/devices/%s/endpoints/%s/cdata?name=%s", deviceID, alarmID, cmd),
		body,
		func(resp *http.Response) {
			data, err := parseAny(resp)
			if err != nil {
				asc.Client.Logger.Error("bad alarm cdata response")

				return
			}

			hook(data)
		},
	)
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request alarm cdata"}
	}

	return nil
}
