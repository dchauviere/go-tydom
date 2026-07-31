package gateway

import (
	"context"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dchauviere/go-tydom/internal/config"
	"github.com/dchauviere/go-tydom/pkg/tydom"
	tydomApi "github.com/dchauviere/go-tydom/pkg/tydom/api"
)

type WebUI struct {
	server         *http.Server
	username       string
	password       string
	tydomAPIClient *tydomApi.ClientAPI
	logger         *slog.Logger
	shutdown       chan struct{}
	gateway        *TydomGateway
	config         *WebUIConfig
}

func (tg *TydomGateway) SetupWebUI(config *WebUIConfig) error {
	if !config.Enabled {
		tg.logger.Info("Web UI disabled")

		return nil
	}

	tg.logger.Info("Web UI enabled")

	if tg.TydomClient == nil {
		tg.logger.Error("tydom client not initialized")

		return fmt.Errorf("tydom client not initialized")
	}

	tg.webUI = &WebUI{
		username:       config.Username,
		password:       config.Password,
		tydomAPIClient: &tydomApi.ClientAPI{Client: tg.TydomClient},
		logger:         slog.With(slog.String("module", "WebUI")),
		shutdown:       make(chan struct{}),
		gateway:        tg,
		config:         config,
	}

	subFS, err := fs.Sub(staticFiles, "ui/dist")
	if err != nil {
		return fmt.Errorf("error loading embedded ui filesystem (%w)", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /api/infos", tg.webUI.basicAuthMiddleware(http.HandlerFunc(tg.webUI.infos)))
	mux.Handle("GET /api/status", tg.webUI.basicAuthMiddleware(http.HandlerFunc(tg.webUI.status)))
	mux.Handle("GET /api/devices", tg.webUI.basicAuthMiddleware(http.HandlerFunc(tg.webUI.devicesHandler)))
	mux.Handle("DELETE /api/device/{id}", tg.webUI.basicAuthMiddleware(http.HandlerFunc(tg.webUI.deleteDevice)))
	mux.Handle("PUT /api/devices/{deviceID}/{endpointID}/name", tg.webUI.basicAuthMiddleware(http.HandlerFunc(tg.webUI.setDeviceName)))
	// Serve embedded files
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	tg.webUI.server = &http.Server{
		Addr:        config.Addr,
		Handler:     mux,
		ReadTimeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return nil
}

//go:embed ui/dist/*.html ui/dist/*.png ui/dist/assets/*
var staticFiles embed.FS

type Device struct {
	DeviceID   int    `json:"deviceId"`
	EndpointID int    `json:"endpointId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}

type Infos struct {
	Version string `json:"version"`
}

type Status struct {
	MQTT  bool `json:"mqtt"`
	Tydom bool `json:"tydom"`
}

func (a *WebUI) infos(resp http.ResponseWriter, _ *http.Request) {
	resp.Header().Add("Content-Type", "application/json")

	infos := Infos{Version: config.VERSION}

	data, err := json.Marshal(infos)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(resp, "Internal error")

		return
	}

	_, _ = resp.Write(data)
}

func (a *WebUI) status(resp http.ResponseWriter, _ *http.Request) {
	resp.Header().Add("Content-Type", "application/json")

	status := Status{
		MQTT:  a.gateway.mqttClient.IsConnected(),
		Tydom: a.gateway.TydomClient.IsConnected(),
	}

	data, err := json.Marshal(status)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(resp, "Internal error")

		return
	}

	_, _ = resp.Write(data)
}

func (a *WebUI) deleteDevice(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")

	deviceID, _ := strconv.Atoi(req.PathValue("id"))

	if err := a.gateway.DeleteDevice(deviceID); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(resp, "Internal error")

		return
	}
	resp.WriteHeader(http.StatusNoContent)
}

func (a *WebUI) setDeviceName(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")

	deviceID, _ := strconv.Atoi(req.PathValue("deviceID"))
	endpointID, _ := strconv.Atoi(req.PathValue("endpointID"))
	a.logger.Info("setting name", "deviceid", deviceID, "endpointid", endpointID)

	body, _ := io.ReadAll(req.Body)
	defer req.Body.Close()

	if err := a.gateway.SetDeviceName(deviceID, endpointID, string(body)); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(resp, "Internal error")

		return
	}

	resp.WriteHeader(http.StatusOK)
	fmt.Fprintln(resp, "Name set")
}

func (a *WebUI) devicesHandler(resp http.ResponseWriter, _ *http.Request) {
	resp.Header().Add("Content-Type", "application/json")

	if a.tydomAPIClient.Client.State() != tydom.Connected {
		resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(resp, "Internal error")

		return
	}

	userConfig, err := a.tydomAPIClient.GetUserConfig()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(resp, "Internal error")

		return
	}

	devicesList := []Device{}

	for _, device := range userConfig.Endpoints {
		devicesList = append(devicesList, Device{
			DeviceID:   device.DeviceID,
			EndpointID: device.EndpointID,
			Name:       device.Name,
			Type:       device.FirstUsage,
		})
	}

	data, err := json.Marshal(devicesList)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(resp, "Internal error")

		return
	}

	_, _ = resp.Write(data)
}

func (a *WebUI) basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Basic ") {
			resp.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(resp, "Unauthorized", http.StatusUnauthorized)

			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[len("Basic "):])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || pair[0] != a.username || pair[1] != a.password {
			http.Error(resp, "Unauthorized", http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(resp, req)
	})
}

func (a *WebUI) Start() {
	go func() {
		a.logger.Info("WebUI listening on " + a.server.Addr)

		if a.config.TLS.Enabled {
			if err := a.server.ListenAndServeTLS(a.config.TLS.CertFile, a.config.TLS.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				a.logger.Error("web server error", "error", err)
			}
		} else {
			if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				a.logger.Error("web server error", "error", err)
			}
		}
	}()
}

func (a *WebUI) Stop() {
	a.logger.Debug("Stopping web server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Server forced to shutdown", "error", err)
	}
}
