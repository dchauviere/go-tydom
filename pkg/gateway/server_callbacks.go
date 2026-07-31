package gateway

import (
	"encoding/json"
	"strings"

	"github.com/dchauviere/go-tydom/pkg/tydom/api"
)

func (tg *TydomGateway) serverHookDeviceInstall(data []byte) {
	var result api.DeviceInstallStatus

	if err := json.Unmarshal(data, &result); err != nil {
		tg.logger.Error("failed to parse json", "error", err)

		return
	}

	if strings.ToLower(result.Protocol) == "x3d" {
		tg.publishAttributes("installStatus", result.InstallStatus)
		if result.InstallStatus == "running" {
			tg.isInstallRunning = true
		} else {
			if tg.isInstallRunning {
				// refresh devices after installation
				if err := tg.tydomAsyncAPI.GetDevicesAccess(tg.hookProcessDevicesAccess); err != nil {
					tg.logger.Error("failed to send devices access request", "error", err)
				}
			}
			tg.isInstallRunning = false
		}
	}
}

func (tg *TydomGateway) serverHookDevicesData(data []byte) {
	var result []api.DevicesData

	if err := json.Unmarshal(data, &result); err != nil {
		tg.logger.Error("failed to parse json", "error", err)

		return
	}

	tg.logger.Debug("devices data received", "data", result)
	tg.updateDevicesData(result)
}

func (tg *TydomGateway) registerServerHooks() {
	tg.TydomClient.AddRequestHook("PUT", "/devices/install", tg.serverHookDeviceInstall)
	tg.TydomClient.AddRequestHook("PUT", "/devices/data", tg.serverHookDevicesData)
}

func (tg *TydomGateway) unregisterServerHooks() {
	tg.TydomClient.RemoveRequestHook("PUT", "/devices/install")
	tg.TydomClient.RemoveRequestHook("PUT", "/devices/data")
}
