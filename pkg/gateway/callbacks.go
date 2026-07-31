package gateway

import (
	"fmt"
	"strings"

	"github.com/dchauviere/go-tydom/internal/helpers"
	"github.com/dchauviere/go-tydom/pkg/tydom"
	tydomAPI "github.com/dchauviere/go-tydom/pkg/tydom/api"
)

func (tg *TydomGateway) hookProcessInfo(data *tydomAPI.Info) {
	tg.logger.Debug("info received", "data", data)

	if tg.State == Starting {
		if err := tg.tydomAPI.SetGatewayAPIMode(); err != nil {
			tg.logger.Error("failed to set api mode")
			return
		}

		err := tg.sendGatewayDiscovery(data)
		if err != nil {
			tg.logger.Error("failed to prepare gateway discovery", "error", err)
		}

		tg.State = Started

		if err := tg.tydomAsyncAPI.GetDevicesAccess(tg.hookProcessDevicesAccess); err != nil {
			tg.logger.Error("failed to send get config order")
		}
	}

	installState := "unknown"

	for _, proto := range data.Protocols {
		if strings.ToLower(proto.Protocol) == "x3d" {
			installState = proto.InstallStatus
		}
	}

	tg.publishAttributes("installState", installState)
	tg.publishAttributes("installType", "None")

	if tg.TydomClient.State() == tydom.Connected {
		tg.publishAttributes("tydomState", "online")
	} else {
		tg.publishAttributes("tydomState", "offline")
	}
}

func findEndpointInUserConfig(userConfig *tydomAPI.UserConfig, deviceID int, endpointID int) *tydomAPI.Endpoint {
	for _, entry := range userConfig.Endpoints {
		if entry.DeviceID == deviceID && entry.EndpointID == endpointID {
			return &entry
		}
	}

	return nil
}

func (tg *TydomGateway) hookProcessDevicesAccess(data *[]tydomAPI.DeviceAccess) {
	if len(*data) == 0 {
		tg.logger.Debug("no device paired")

		return
	}

	gwInfo, err := tg.tydomAPI.GetInfo()
	if err != nil {
		tg.logger.Error("failed to get gateway infos", "error", err)

		return
	}

	cfgResp, err := tg.tydomAPI.GetUserConfig()
	if err != nil {
		tg.logger.Error("failed to get user configs", "error", err)

		return
	}

	for _, device := range *data {
		for _, endpoint := range device.Endpoints {
			name := fmt.Sprintf(
				"%s %d-%d",
				helpers.CapitalizeFirstLetterUnicode(endpoint.Access.Profile),
				device.ID,
				endpoint.ID,
			)

			if cfgEndpoint := findEndpointInUserConfig(cfgResp, device.ID, endpoint.ID); cfgEndpoint == nil {
				cfgResp.Endpoints = append(cfgResp.Endpoints, tydomAPI.Endpoint{
					DeviceID:   device.ID,
					EndpointID: endpoint.ID,
					FirstUsage: endpoint.Access.Profile,
					//Skill:             "",
					Name:              name,
					AnticipationStart: false,
					Picto:             endpoint.Access.Profile,
					LastUsage:         endpoint.Access.Profile,
					//WidgetBehaviour:   tydomAPI.WidgetBehaviour{},
				})
				cfgResp.CatalogID = "F2BD90F93B888DA02C54980F11AE4796DFCC98F447CD3FE326F5A3A964C939BF"
				if _, err := tg.tydomAPI.SetUserConfig(cfgResp); err != nil {
					tg.logger.Error("failed to store configs", "error", err)

					return
				}
			} else {
				name = cfgEndpoint.Name
			}

			if err = tg.sendDeviceDiscovery(gwInfo, device.ID, endpoint.ID, name, endpoint.Access.Profile); err != nil {
				tg.logger.Error("failed to send discovery", "deviceID", device.ID, "endpointID", endpoint.ID)
			}
		}
	}
}

func (tg *TydomGateway) hookProcessDevicesData(data *[]tydomAPI.DevicesData) {
	tg.logger.Debug("devices data received", "data", data)
	tg.updateDevicesData(*data)
}

func (tg *TydomGateway) tydomDispatchWrite(deviceID, endpointID, paramName, paramValue string) {
	tg.logger.Debug(
		"dispatch data to tydom",
		"deviceID", deviceID,
		"endpointID", endpointID,
		"paramName", paramName,
		"paramValue", paramValue,
	)

	_, err := tg.tydomAPI.SetDeviceData(deviceID, endpointID, tydomAPI.KeyValue{Name: paramName, Value: paramValue})
	if err != nil {
		tg.logger.Error(
			"failed to write tydom data",
			"deviceID", deviceID,
			"endpointID", endpointID,
			"paramName", paramName,
			"paramValue", paramValue,
			"error", err,
		)
	}
}
