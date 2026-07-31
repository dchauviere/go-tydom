package api

import (
	"fmt"
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

var DeviceNotFoundError = &tydom.Error{Msg: "device not found"}

type DeviceInstallStatus struct {
	Protocol      string `json:"protocol"`
	InstallStatus string `json:"installStatus"`
}

func parseDeviceInstallStatus(resp *http.Response) (*DeviceInstallStatus, error) {
	var data DeviceInstallStatus

	err := parse(resp, &data)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to parse device install status response"}
	}

	return &data, nil
}

// Set device in permit join.
//
// data => {"protocol":"X3D","profile":"shutter","type":"x3d_rm","net":0}
// data => {"protocol":"X3D","profile":"thermic","type":"x3d_rm","net":4}
// POST /devices/install {"protocol":"X3D","profile":"thermic","type":"x3d_rm","net":4}
// PUT /devices/install{"protocol":"X3D","installStatus":"running"}
// PUT /devices/install{"protocol":"X3D","installStatus":"idle"}
// POST /devices/install {"protocol":"X3D","profile":"shutter","type":"x3d_rm","net":0}
// PUT /devices/install{"protocol":"X3D","installStatus":"running"}
// PUT /devices/install{"protocol":"X3D","installStatus":"idle"}.
func (ac *ClientAPI) SetDevicesInstall(protocol, profile, installType string, net int) (*DeviceInstallStatus, error) {
	resp, err := ac.Client.HTTPRequest(
		"POST",
		"/devices/install",
		fmt.Sprintf(`{"protocol":"%s","profile":"%s","type":"%s","net":%d}`, protocol, profile, installType, net),
	)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send device install request"}
	}

	return parseDeviceInstallStatus(resp)
}

func (asc *ClientAsyncAPI) SetDevicesInstall(
	protocol, profile, installType string, net int, hook func(data *DeviceInstallStatus),
) error {
	err := asc.Client.HTTPAsyncRequest(
		"POST",
		"/devices/install",
		fmt.Sprintf(`{"protocol":"%s","profile":"%s","type":"%s","net":%d}`, protocol, profile, installType, net),
		func(resp *http.Response) {
			data, err := parseDeviceInstallStatus(resp)
			if err != nil {
				asc.Client.Logger.Error("bad device install response")

				return
			}

			hook(data)
		},
	)
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request device install"}
	}

	return nil
}

// Get Devices Install Status.
// return {"protocol":"X3D","installStatus":"idle"} or {"protocol":"X3D","installStatus":"running"}.
func (ac *ClientAPI) GetDevicesInstallStatus() (*DeviceInstallStatus, error) {
	resp, err := ac.Client.HTTPRequest("PUT", "/devices/install", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send Info request"}
	}

	return parseDeviceInstallStatus(resp)
}

func (asc *ClientAsyncAPI) GetDevicesInstallStatus(hook func(data *DeviceInstallStatus)) error {
	err := asc.Client.HTTPAsyncRequest("PUT", "/devices/install", "", func(resp *http.Response) {
		data, err := parseDeviceInstallStatus(resp)
		if err != nil {
			asc.Client.Logger.Error("bad device install response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request device install"}
	}

	return nil
}

// Delete a device.
func (ac *ClientAPI) DeleteDevice(deviceID int) (*any, error) {
	resp, err := ac.Client.HTTPRequest("DELETE", fmt.Sprintf("/devices/%d", deviceID), "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send delete device request"}
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		// Device deleted successfully
		return parseAny(resp)
	case http.StatusNotFound:
		return nil, DeviceNotFoundError
	default:
		return nil, &tydom.Error{Msg: fmt.Sprintf("failed to delete device, status code: %d", resp.StatusCode)}
	}
}

func (asc *ClientAsyncAPI) DeleteDevice(deviceID int, hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("DELETE", fmt.Sprintf("/devices/%d", deviceID), "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad delete device response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request delete device"}
	}

	return nil
}
