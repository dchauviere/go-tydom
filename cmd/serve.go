/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dchauviere/go-tydom/pkg/gateway"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type serveCommand struct{}

func (sc *serveCommand) Command() *cobra.Command {

	// serveCmd represents the serve command
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			if viper.GetBool("webui.tls.enabled") && (viper.GetString("webui.tls.certfile") == "" || viper.GetString("webui.tls.keyfile") == "") {
				slog.Error("certfile and keyfile are mandatory for TLS")

				return
			}

			tydomGateway, err := gateway.NewTydomGateway(&gateway.Config{
				MQTT: &gateway.MQTTConfig{
					Enabled:              viper.GetBool("mqtt.enabled"),
					URL:                  viper.GetString("mqtt.url"),
					ClientID:             viper.GetString("mqtt.clientid"),
					User:                 viper.GetString("mqtt.username"),
					Password:             viper.GetString("mqtt.password"),
					CleanSession:         viper.GetBool("mqtt.cleanSession"),
					Store:                viper.GetString("mqtt.store"),
					DiscoveryTopicPrefix: viper.GetString("mqtt.discoveryTopicPrefix"),
					GatewayTopicPrefix:   viper.GetString("mqtt.gatewayTopicPrefix"),
					CAFile:               viper.GetString("mqtt.cafile"),
				},
				Tydom: &gateway.TydomConfig{
					GatewayID: viper.GetString("tydom.gateway-id"),
					Hostname:  viper.GetString("tydom.hostname"),
					Password:  viper.GetString("tydom.password"),
				},
				WebUI: &gateway.WebUIConfig{
					Enabled:  viper.GetBool("webui.enabled"),
					Addr:     viper.GetString("webui.listen"),
					Username: viper.GetString("webui.username"),
					Password: viper.GetString("webui.password"),
					TLS: &gateway.WebUITLSConfig{
						Enabled:  viper.GetBool("webui.tls.enabled"),
						CertFile: viper.GetString("webui.tls.certfile"),
						KeyFile:  viper.GetString("webui.tls.keyfile"),
					},
				},
			})
			if err != nil {
				slog.Error("failed to setup gateway", "error", err)

				return
			}

			defer tydomGateway.Stop()

			if err := tydomGateway.Start(); err != nil {
				slog.Error("failed to start gateway", "error", err)

				return
			}

			terminateSignals := make(chan os.Signal, 1)
			reloadSignals := make(chan os.Signal, 1)
			signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM)
			signal.Notify(reloadSignals, syscall.SIGUSR1)
			for { // We are looping here because config reload can happen multiple times.
				select {
				case <-terminateSignals:
					slog.Info("shutting down server gracefully")

					return
				case <-reloadSignals:
					slog.Info("Got reload signal, will reload")
					// Some config reload code here.
				}
			}
		},
	}
	sc.init(serveCmd)
	return serveCmd
}

func (sc *serveCommand) init(cmd *cobra.Command) {
	cmd.Flags().Bool("webui", true, "WebUI enabled")
	_ = viper.BindPFlag("webui.enabled", cmd.Flags().Lookup("webui"))

	cmd.Flags().String("webui-listen", "0.0.0.0:8080", "WebUI listen address")
	_ = viper.BindPFlag("webui.listen", cmd.Flags().Lookup("webui-listen"))

	cmd.Flags().String("webui-username", "admin", "WebUI basic auth username")
	_ = viper.BindPFlag("webui.username", cmd.Flags().Lookup("webui-username"))

	cmd.Flags().String("webui-password", "admin", "WebUI basic auth password")
	_ = viper.BindPFlag("webui.password", cmd.Flags().Lookup("webui-password"))

	cmd.Flags().Bool("webui-tls", false, "WebUI enable TLS")
	_ = viper.BindPFlag("webui.tls.enabled", cmd.Flags().Lookup("webui-tls"))

	cmd.Flags().String("webui-tls-certfile", "", "WebUI TLS certificate file")
	_ = viper.BindPFlag("webui.tls.certfile", cmd.Flags().Lookup("webui-tls-certfile"))

	cmd.Flags().String("webui-tls-keyfile", "", "WebUI TLS key file")
	_ = viper.BindPFlag("webui.tls.keyfile", cmd.Flags().Lookup("webui-tls-keyfile"))

	cmd.Flags().Bool("mqtt", true, "MQTT enabled")
	_ = viper.BindPFlag("mqtt.enabled", cmd.Flags().Lookup("mqtt"))

	cmd.Flags().String("mqtt-url", "tcp://localhost:1883", "MQTT url to connect")
	_ = viper.BindPFlag("mqtt.url", cmd.Flags().Lookup("mqtt-url"))

	cmd.Flags().String("mqtt-client-id", "go-tydom", "MQTT client ID")
	_ = viper.BindPFlag("mqtt.clientid", cmd.Flags().Lookup("mqtt-client-id"))

	cmd.Flags().String("mqtt-username", "", "MQTT Username")
	_ = viper.BindPFlag("mqtt.username", cmd.Flags().Lookup("mqtt-username"))

	cmd.Flags().String("mqtt-password", "", "MQTT Password")
	_ = viper.BindPFlag("mqtt.password", cmd.Flags().Lookup("mqtt-password"))

	cmd.Flags().String("mqtt-store", ":memory:", "MQTT store")
	_ = viper.BindPFlag("mqtt.store", cmd.Flags().Lookup("mqtt-store"))

	cmd.Flags().String("mqtt-cafile", "", "MQTT CA file")
	_ = viper.BindPFlag("mqtt.cafile", cmd.Flags().Lookup("mqtt-cafile"))

	cmd.Flags().BoolP("mqtt-clean-session", "m", false, "toggle MQTT clean session")
	_ = viper.BindPFlag("mqtt.cleanSession", cmd.Flags().Lookup("mqtt-clean-session"))

	cmd.Flags().String(
		"mqtt-discovery-topic-prefix",
		"homeassistant",
		"MQTT discovery topic prefix",
	)
	_ = viper.BindPFlag("mqtt.discoveryTopicPrefix", cmd.Flags().Lookup("mqtt-discovery-topic-prefix"))

	cmd.Flags().String("gateway-topic-prefix", "gt2m", "MQTT topic prefix")
	_ = viper.BindPFlag("mqtt.gatewayTopicPrefix", cmd.Flags().Lookup("gateway-topic-prefix"))
}
