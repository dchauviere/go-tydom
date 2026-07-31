package gateway

import (
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/dchauviere/go-tydom/internal/config"
	tydomAPI "github.com/dchauviere/go-tydom/pkg/tydom/api"
)

//go:embed gateway.json
var gatewayDiscoveryPayload string

//go:embed shutter.json
var shutterDiscoveryPayload string

type discoveryInfo struct {
	MQTTPrefix    string
	ID            string
	Model         string
	HWVersion     string
	SWVersion     string
	Manufacturer  string
	AppVersion    string
	AppSupportURL string
	AppName       string
	InstallCodes  string
}

type deviceDiscoveryInfo struct {
	Gateway     discoveryInfo
	DeviceID    string
	EndpointID  string
	DeviceName  string
	DeviceModel string
	Profile     string
}

func buildGatewayDiscoveryPayload(infos *discoveryInfo) (string, error) {
	tmpl, err := template.New("test").Parse(gatewayDiscoveryPayload)
	if err != nil {
		return "", fmt.Errorf("failed to parse gateway discovery template (%w)", err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, infos)
	if err != nil {
		return "", fmt.Errorf("failed to process gateway discovery template (%w)", err)
	}

	return buf.String(), nil
}

func buildDeviceDiscoveryPayload(infos *deviceDiscoveryInfo) (string, error) {
	tmpl, err := template.New("test").Parse(shutterDiscoveryPayload)
	if err != nil {
		return "", fmt.Errorf("failed to parse shutter discovery template (%w)", err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, infos)
	if err != nil {
		return "", fmt.Errorf("failed to process shutter discovery template (%w)", err)
	}

	return buf.String(), nil
}

func (tg *TydomGateway) sendGatewayDiscovery(gwInfo *tydomAPI.Info) error {
	tg.logger.Debug("sending gateway discovery")

	installCodesNames := defaultDevicesInstallCodes.GetNames()
	installCodesNames = append([]string{"None"}, installCodesNames...)

	payload, err := buildGatewayDiscoveryPayload(&discoveryInfo{
		MQTTPrefix:    tg.topicPrefix,
		ID:            tg.gatewayID,
		Model:         gwInfo.ProductName,
		HWVersion:     gwInfo.MainVersionHw,
		SWVersion:     gwInfo.MainVersionSw,
		Manufacturer:  "Delta Dore",
		AppVersion:    config.VERSION,
		AppName:       config.USERAGENT,
		AppSupportURL: config.SUPPORT_URL,
		InstallCodes:  strings.Join(installCodesNames, "\",\""),
	})
	if err != nil {
		return fmt.Errorf("failed to build gateway discovery payload (%w)", err)
	}

	tg.mqttClient.Publish(fmt.Sprintf("%s/device/%s/config", tg.discoveryPrefix, tg.uniqueIDPrefix), 1, true, payload)

	return nil
}

func (tg *TydomGateway) sendDeviceDiscovery(
	gwInfo *tydomAPI.Info,
	deviceID, endpointID int,
	name, profile string,
) error {
	tg.logger.Debug("sending device discovery")

	payload, err := buildDeviceDiscoveryPayload(
		&deviceDiscoveryInfo{
			Gateway: discoveryInfo{
				MQTTPrefix:    tg.topicPrefix,
				ID:            tg.gatewayID,
				Model:         gwInfo.ProductName,
				HWVersion:     gwInfo.MainVersionHw,
				SWVersion:     gwInfo.MainVersionSw,
				Manufacturer:  "Delta Dore",
				AppVersion:    config.VERSION,
				AppName:       config.USERAGENT,
				AppSupportURL: config.SUPPORT_URL,
			},
			DeviceID:   strconv.Itoa(deviceID),
			EndpointID: strconv.Itoa(endpointID),
			DeviceName: name,
			Profile:    profile,
		})
	if err != nil {
		return fmt.Errorf("failed to build device discovery payload (%w)", err)
	}

	topic := fmt.Sprintf(
		"%s/device/%s/%s/config",
		tg.discoveryPrefix,
		tg.uniqueIDPrefix,
		fmt.Sprintf("%s-%d-%d", tg.uniqueIDPrefix, deviceID, endpointID),
	)

	tg.logger.Debug("sending device discovery", "data", payload)
	tg.mqttClient.Publish(topic, 1, true, payload)

	return nil
}
