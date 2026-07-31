package gateway

import (
	"fmt"
	"strings"

	"github.com/dchauviere/go-tydom/internal/logging"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type DeviceInstallCodes struct {
	Name     string
	Protocol string `json:"protocol"`
	Profile  string `json:"profile"`
	Type     string `json:"type"`
	Net      int    `json:"net"`
}

type DevicesInstallCodes []DeviceInstallCodes

func (dic *DevicesInstallCodes) GetNames() []string {
	names := make([]string, 0)

	for _, elem := range *dic {
		names = append(names, elem.Name)
	}

	return names
}

func (dic *DevicesInstallCodes) GetInstallCodes(name string) (*DeviceInstallCodes, error) {
	for _, elem := range *dic {
		if elem.Name == name {
			return &elem, nil
		}
	}

	return nil, fmt.Errorf("install codes not found for name")
}

var defaultDevicesInstallCodes = DevicesInstallCodes{
	{Name: "Shutter", Protocol: "X3D", Profile: "shutter", Type: "x3d_rm", Net: 0},
	{Name: "Thermic", Protocol: "X3D", Profile: "thermic", Type: "x3d_rm", Net: 4},
}

type DeviceState int

const (
	Offline DeviceState = iota
	Online
)

func (d DeviceState) String() string {
	switch d {
	case Offline:
		return "offline"
	case Online:
		return "online"
	default:
		return "unknown"
	}
}

type Command struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"params"`
}

type CommandInstallParams struct {
	Protocol string `json:"protocol"`
	Profile  string `json:"profile"`
	Type     string `json:"type"`
	Net      int    `json:"net"`
}

func (tg *TydomGateway) processGatewayInstallCommand(name string) {
	tg.logger.Info("set device install", "name", name)

	if strings.ToLower(name) == "none" {
		return
	}

	installCodes, err := defaultDevicesInstallCodes.GetInstallCodes(name)
	if err != nil {
		tg.logger.Error("install codes not found", "name", name)
	}

	result, err := tg.tydomAPI.SetDevicesInstall(
		installCodes.Protocol,
		installCodes.Profile,
		installCodes.Type,
		installCodes.Net,
	)
	if err != nil {
		tg.logger.Error("failed to send install command to tydom")

		return
	}

	tg.logger.Debug("set device install result", "result", result)
	tg.publishAttributes("installState", result.InstallStatus)
	tg.publishAttributes("installType", "None")
}

func (tg *TydomGateway) processGatewayCommand(_ MQTT.Client, msg MQTT.Message) {
	tg.logger.Debug("gateway command received", "payload", msg.Payload())

	topicSplit := strings.Split(msg.Topic(), "/")
	cmd := strings.ToLower(topicSplit[len(topicSplit)-1])

	switch cmd {
	case "loglevel":
		tg.logger.Info("setting log level", "level", msg.Payload())

		_ = logging.SetLevel(string(msg.Payload()))

		tg.publishAttributes("logLevel", logging.Level())
	case "installtype":
		tg.logger.Info("installing new device", "device", msg.Payload())
		tg.processGatewayInstallCommand(string(msg.Payload()))
	default:
		tg.logger.Error("unknown gateway command", "cmd", cmd)
	}
}
