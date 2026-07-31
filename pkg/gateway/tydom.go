package gateway

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dchauviere/go-tydom/pkg/tydom"
	tydomAPI "github.com/dchauviere/go-tydom/pkg/tydom/api"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func (tg *TydomGateway) SetupTydom(config *TydomConfig) error {
	tg.TydomClient = tydom.NewClient(config.Hostname, config.GatewayID, config.Password)
	tg.tydomAPI = &tydomAPI.ClientAPI{Client: tg.TydomClient}
	tg.tydomAsyncAPI = &tydomAPI.ClientAsyncAPI{Client: tg.TydomClient}
	tg.gatewayID = strings.ToLower(tg.TydomClient.GatewayID)
	tg.uniqueIDPrefix = "gt2m_" + strings.ToLower(tg.TydomClient.GatewayID)

	return nil
}

func (tg *TydomGateway) updateDevicesData(devicesData []tydomAPI.DevicesData) {
	for _, data := range devicesData {
		for _, endpoint := range data.Endpoints {
			for _, entry := range endpoint.Data {
				// Publish to mqtt gt2m/<device_id>/<endpoint>/params/<name>
				var paramValue string

				paramName, valid := entry["name"].(string)

				if !valid {
					tg.logger.Error("failed to get param name", "device_id", data.ID, "endpoint_id", endpoint.ID, "entry", entry)

					continue
				}

				switch value := entry["value"].(type) {
				case string:
					paramValue = value
				case bool:
					paramValue = strconv.FormatBool(value)
				case int, int8, int16, int32, int64:
					paramValue = fmt.Sprintf("%d", value)
				case float32, float64:
					paramValue = fmt.Sprintf("%.2f", value)
				}

				topic := fmt.Sprintf("%s/%s/devices/%d/%d/attrs/%s", tg.topicPrefix, tg.gatewayID, data.ID, endpoint.ID, paramName)

				tg.logger.Debug("update data to mqtt", "name", paramName, "value", entry["value"], "topic", topic)
				tg.mqttClient.Publish(topic, 1, true, paramValue)
			}
		}
	}
}

func (tg *TydomGateway) processDiscovery(_ MQTT.Client, msg MQTT.Message) {
	if string(msg.Payload()) == "online" {
		infos, err := tg.tydomAPI.GetInfo()
		if err != nil {
			tg.logger.Error("failed to get tydom info for sending discovery")

			return
		}

		if err = tg.sendGatewayDiscovery(infos); err != nil {
			tg.logger.Error("failed to send discovery")
		}
	}
}

func (tg *TydomGateway) processCommand(_ MQTT.Client, msg MQTT.Message) {
	tg.logger.Info("receive mqtt command", "topic", msg.Topic(), "payload", msg.Payload())
	rgxp := regexp.MustCompile(
		fmt.Sprintf(
			"%s/%s/devices/(?P<deviceID>.+)/(?P<endpointID>.+)/set/(?P<paramName>.+)",
			tg.topicPrefix,
			tg.gatewayID,
		),
	)
	res := rgxp.FindStringSubmatch(msg.Topic())

	if res != nil {
		tg.logger.Debug(
			"mqtt command",
			"deviceID",
			res[rgxp.SubexpIndex("deviceID")],
			"endpointID",
			res[rgxp.SubexpIndex("endpointID")],
			"paramName",
			res[rgxp.SubexpIndex("paramName")],
		)
	}

	tg.tydomDispatchWrite(
		res[rgxp.SubexpIndex("deviceID")],
		res[rgxp.SubexpIndex("endpointID")],
		res[rgxp.SubexpIndex("paramName")],
		string(msg.Payload()),
	)
}
