package gateway

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"time"

	"github.com/dchauviere/go-tydom/internal/logging"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (tg *TydomGateway) publishAttributes(attribute string, value string) {
	if token := tg.mqttClient.Publish(
		tg.topicPrefix+"/"+tg.gatewayID+"/attrs/"+attribute,
		1,
		true,
		value,
	); token.Wait() && token.Error() != nil {
		tg.logger.Error("failed to publish attribute", "error", token.Error())
	}
}

func (tg *TydomGateway) publishAvailability(state DeviceState) {
	if token := tg.mqttClient.Publish(
		tg.topicPrefix+"/"+tg.gatewayID+"/availability",
		1,
		true,
		state.String(),
	); token.Wait() && token.Error() != nil {
		tg.logger.Error("failed to publish lwt message to mqtt", "error", token.Error())
	}
}

func (tg *TydomGateway) SetupMQTT(config *MQTTConfig) error {
	if !config.Enabled {
		tg.logger.Info("MQTT disabled")

		return nil
	}

	tg.logger.Info("MQTT enabled")

	tg.topicPrefix = config.GatewayTopicPrefix
	tg.discoveryPrefix = config.DiscoveryTopicPrefix

	mqttOptions := mqtt.NewClientOptions()
	mqttOptions.AddBroker(config.URL)
	mqttOptions.SetClientID(config.ClientID)
	mqttOptions.SetUsername(config.User)
	mqttOptions.SetPassword(config.Password)
	mqttOptions.SetCleanSession(config.CleanSession)
	mqttOptions.SetConnectRetry(true)
	mqttOptions.SetConnectRetryInterval(5 * time.Second)
	mqttOptions.SetAutoReconnect(true)

	if config.Store != ":memory:" {
		mqttOptions.SetStore(mqtt.NewFileStore(config.Store))
	}

	if config.CAFile != "" {
		certpool := x509.NewCertPool()
		ca, err := os.ReadFile(config.CAFile)
		if err != nil {
			tg.logger.Error("failed to read cafile", "error", err)
			return err
		}
		certpool.AppendCertsFromPEM(ca)
		mqttOptions.SetTLSConfig(&tls.Config{
			RootCAs: certpool,
		})
	}

	mqttOptions.SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
		tg.logger.Debug("message received", "topic", msg.Topic(), "payload", msg.Payload())
	})

	mqttOptions.SetWill(tg.topicPrefix+"/"+tg.gatewayID+"/availability", "offline", 1, true)
	mqttOptions.SetOnConnectHandler(func(mqttClient mqtt.Client) {
		tg.logger.Info("MQTT connected")

		if token := mqttClient.Subscribe(
			tg.topicPrefix+"/"+tg.gatewayID+"/set/+",
			0,
			tg.processGatewayCommand,
		); token.Wait() && token.Error() != nil {
			tg.logger.Error("failed to subscribe to mqtt", "error", token.Error())
		}

		if token := tg.mqttClient.Subscribe(
			tg.topicPrefix+"/"+tg.gatewayID+"/devices/+/+/set/+",
			0,
			tg.processCommand,
		); token.Wait() && token.Error() != nil {
			tg.logger.Error("failed to subscribe to mqtt", "error", token.Error())
		}

		if token := tg.mqttClient.Subscribe(
			tg.discoveryPrefix+"/status",
			0,
			tg.processDiscovery,
		); token.Wait() && token.Error() != nil {
			tg.logger.Error("failed to subscribe to mqtt", "error", token.Error())
		}

		tg.publishAvailability(Online)
		tg.publishAttributes("logLevel", logging.Level())
	})

	tg.mqttClient = mqtt.NewClient(mqttOptions)

	return nil
}

func (tg *TydomGateway) mqttConnect() {
	if token := tg.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		tg.logger.Error("failed to connect to mqtt", "error", token.Error())
	}
}

func (tg *TydomGateway) mqttDisconnect() {
	tg.logger.Debug("publishing mqtt will message")

	tg.publishAvailability(Offline)

	tg.logger.Debug("disconnecting mqtt")
	tg.mqttClient.Disconnect(250)
	tg.logger.Info("MQTT disconnected")
}
