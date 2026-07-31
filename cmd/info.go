/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/dchauviere/go-tydom/pkg/tydom"
	tydomAPI "github.com/dchauviere/go-tydom/pkg/tydom/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type infoCmd struct{}

func (ic *infoCmd) Command() *cobra.Command {
	// infoCmd represents the info command
	var infoCmd = &cobra.Command{
		Use:   "info",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			myclient := tydom.NewClient(
				viper.GetString("tydom.hostname"),
				viper.GetString("tydom.gateway-id"),
				viper.GetString("tydom.password"),
			)
			tydomClientAPI := tydomAPI.ClientAPI{Client: myclient}
			if err := myclient.Start(); err != nil {
				slog.Error("failed to start tydom client", "error", err)
				os.Exit(1)
			}
			defer myclient.Stop()

			result, err := tydomClientAPI.GetInfo()
			if err != nil {
				slog.Error("error sending command", "error", err)
				return
			}
			fmt.Println(result)
		},
	}

	return infoCmd
}
