package api

import (
	"fmt"
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

type PasswordResponse any

// Set a new password.
func (ac *ClientAPI) SetPassword(newPassword string, oldPassword string) (*any, error) {
	resp, err := ac.Client.HTTPRequest(
		"PUT",
		"/configs/gateway/password",
		fmt.Sprintf("{\"new\":\"%s\",\"old\":\"%s\"}", newPassword, oldPassword),
	)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send password request: " + err.Error()}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) SetPassword(newPassword string, oldPassword string, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest(
		"PUT",
		"/configs/gateway/password",
		fmt.Sprintf("{\"new\":\"%s\",\"old\":\"%s\"}", newPassword, oldPassword),
		func(resp *http.Response) {
			data, err := parseAny(resp)
			if err != nil {
				asc.Client.Logger.Error("bad password response")

				return
			}

			hook(data)
		})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request password"}
	}

	return nil
}
