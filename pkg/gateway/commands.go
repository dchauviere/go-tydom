package gateway

import (
	"fmt"

	"github.com/dchauviere/go-tydom/pkg/tydom/api"
)

func (tg *TydomGateway) processUpdateUserConfig(handler func(userConfig *api.UserConfig) *api.UserConfig) error {
	userConfig, err := tg.tydomAPI.GetUserConfig()
	if err != nil {
		return err
	}

	updatedConfig := handler(userConfig)

	_, err = tg.tydomAPI.SetUserConfig(updatedConfig)
	if err != nil {
		return err
	}
	return nil
}

func (tg *TydomGateway) DeleteDevice(deviceID int) error {
	tg.logger.Info("Deleting device", "device_id", deviceID)
	// First, send the delete request to the Tydom API
	_, err := tg.tydomAPI.DeleteDevice(deviceID)
	if err != nil && err != api.DeviceNotFoundError {
		tg.logger.Error("Failed to delete device from Tydom API", "error", err)
		return err
	}

	// Next, update the user configuration to remove references to the deleted device
	return tg.processUpdateUserConfig(func(userConfig *api.UserConfig) *api.UserConfig {
		newEndpoints := make([]api.Endpoint, 0, len(userConfig.Endpoints))
		for _, endpoint := range userConfig.Endpoints {
			tg.logger.Info("device", "deviceid", endpoint.DeviceID, "endpointid", endpoint.EndpointID)

			if endpoint.DeviceID != deviceID {
				newEndpoints = append(newEndpoints, endpoint)
			} else {
				tg.logger.Debug("cleanup device discovery", "deviceid", endpoint.DeviceID, "endpointid", endpoint.EndpointID)
				tg.mqttClient.Publish(
					fmt.Sprintf(
						"%s/device/%s/%s/config",
						tg.discoveryPrefix,
						tg.uniqueIDPrefix,
						fmt.Sprintf("%s-%d-%d", tg.uniqueIDPrefix, endpoint.DeviceID, endpoint.EndpointID),
					), 1, true, "")
			}
		}
		userConfig.Endpoints = newEndpoints
		return userConfig
	})
}

func (tg *TydomGateway) SetDeviceName(deviceID int, endpointID int, name string) error {
	tg.logger.Info("Setting device name", "device_id", deviceID, "endpoint_id", endpointID, "name", name)
	err := tg.processUpdateUserConfig(func(userConfig *api.UserConfig) *api.UserConfig {
		for index, endpoint := range userConfig.Endpoints {
			tg.logger.Info("device", "deviceid", endpoint.DeviceID, "endpointid", endpoint.EndpointID)

			if endpoint.DeviceID == deviceID && endpoint.EndpointID == endpointID {
				userConfig.Endpoints[index].Name = name
				break
			}
		}
		return userConfig
	})

	tg.tydomAsyncAPI.GetDevicesAccess(tg.hookProcessDevicesAccess)
	return err
}
