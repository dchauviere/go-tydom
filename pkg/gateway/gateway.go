package gateway

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dchauviere/go-tydom/pkg/tydom"
	tydomAPI "github.com/dchauviere/go-tydom/pkg/tydom/api"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type State int

const (
	Stopped State = iota
	Starting
	Started
)

func (s State) String() string {
	switch s {
	case Stopped:
		return "Stopped"
	case Starting:
		return "Starting"
	case Started:
		return "Started"
	default:
		return "Unknown"
	}
}

const DefaultPollingInterval int = 15

type TydomGateway struct {
	TydomClient      *tydom.Client
	discoveryPrefix  string
	logger           *slog.Logger
	shutdown         chan struct{}
	mqttDataReceiver chan [2]string
	State            State
	mqttClient       MQTT.Client
	webUI            *WebUI
	tydomAPI         *tydomAPI.ClientAPI
	tydomAsyncAPI    *tydomAPI.ClientAsyncAPI
	gatewayID        string
	uniqueIDPrefix   string
	pollingInterval  int
	topicPrefix      string
	wg               sync.WaitGroup
	isInstallRunning bool
}

/*
NewTydomGateway create a new TydomGateway.
*/
func NewTydomGateway(config *Config) (*TydomGateway, error) {
	tgw := &TydomGateway{
		mqttDataReceiver: make(chan [2]string),
		State:            Stopped,
		logger:           slog.With(slog.String("module", "gateway")),
		pollingInterval:  DefaultPollingInterval,
		shutdown:         make(chan struct{}),
	}

	if err := tgw.SetupTydom(config.Tydom); err != nil {
		return nil, fmt.Errorf("failed to setup Tydom client (%w)", err)
	}

	if err := tgw.SetupMQTT(config.MQTT); err != nil {
		return nil, fmt.Errorf("failed to setup MQTT client (%w)", err)
	}

	if err := tgw.SetupWebUI(config.WebUI); err != nil {
		return nil, fmt.Errorf("failed to setup Web UI (%w)", err)
	}

	return tgw, nil
}

func (tg *TydomGateway) GetUniqueID() string {
	return "tydom-" + strings.ToLower(tg.TydomClient.GatewayID)
}

func (tg *TydomGateway) Start() error {
	if tg.State == Started {
		return nil
	}

	tg.logger.Debug("starting gateway")
	tg.State = Starting

	tg.registerServerHooks()

	// start mqtt if needed
	if tg.mqttClient != nil {
		tg.mqttConnect()
	}

	if err := tg.TydomClient.Start(); err != nil {
		return fmt.Errorf("failed to start tydom client (%w)", err)
	}

	if tg.webUI != nil {
		tg.webUI.Start()
	}

	tg.wg.Add(1)
	go tg.run()

	tg.logger.Debug("Gateway started")

	return nil
}

func (tg *TydomGateway) Stop() {
	tg.logger.Debug("stopping gateway")

	if tg.webUI != nil {
		tg.webUI.Stop()
	}

	tg.logger.Debug("stopping tydom client")
	if tg.TydomClient != nil {
		tg.TydomClient.Stop()
	}

	tg.logger.Debug("stopping gateway goroutines")
	close(tg.shutdown)
	tg.wg.Wait()

	tg.unregisterServerHooks()

	if tg.mqttClient != nil {
		tg.mqttDisconnect()
	}

	tg.logger.Debug("Gateway stopped")
}

func (tg *TydomGateway) GetState() State {
	return tg.State
}

func (tg *TydomGateway) run() {
	defer tg.wg.Done()
	tg.logger.Debug("enter run loop")
	pollingTrigger := time.NewTicker(time.Duration(tg.pollingInterval) * time.Second)

	for {
		select {
		// Process shutdown signal
		case <-tg.shutdown:
			tg.logger.Debug("run shutdown")

			return
		case <-pollingTrigger.C:
			if tg.TydomClient.State() != tydom.Connected {
				tg.logger.Debug("skip updating")

				continue
			}

			tg.logger.Debug("updating gateway state")

			if err := tg.tydomAsyncAPI.GetInfo(tg.hookProcessInfo); err != nil {
				tg.logger.Error("failed to get info", "error", err)
			}

			if err := tg.tydomAsyncAPI.RefreshAll(func(status int) {
				if status != http.StatusOK {
					tg.logger.Error("refresh failed", "http_code", status)
				}
			}); err != nil {
				tg.logger.Error("failed to send refresh", "error", err)
			}

			if err := tg.tydomAsyncAPI.GetDevicesData(tg.hookProcessDevicesData); err != nil {
				tg.logger.Error("failed to get devices data", "error", err)
			}
		}
	}
}
