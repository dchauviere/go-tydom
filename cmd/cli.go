/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/dchauviere/go-tydom/pkg/tydom"
	tydomAPI "github.com/dchauviere/go-tydom/pkg/tydom/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}

	return r, true
}

type cliCmd struct{}

func (cc *cliCmd) Command() *cobra.Command {
	// cliCmd represents the cli command
	var cliCmd = &cobra.Command{
		Use:   "cli",
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

			l, err := readline.NewEx(&readline.Config{
				Prompt:      "\033[31m»\033[0m ",
				HistoryFile: "/tmp/readline.tmp",
				AutoComplete: readline.NewPrefixCompleter(
					readline.PcItem("info"),
					readline.PcItem("password"),
					readline.PcItem("access"),
					readline.PcItem("get"),
					readline.PcItem("post"),
					readline.PcItem("delete"),
				),
				InterruptPrompt: "^C",
				EOFPrompt:       "exit",

				HistorySearchFold:   true,
				FuncFilterInputRune: filterInput,
			})
			if err != nil {
				panic(err)
			}
			defer l.Close()
			l.CaptureExitSignal()

			setPasswordCfg := l.GenPasswordConfig()
			setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
				l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
				l.Refresh()
				return nil, 0, false
			})

			var result any

			for {
				line, err := l.Readline()
				if err == readline.ErrInterrupt {
					if len(line) == 0 {
						break
					} else {
						continue
					}
				} else if err == io.EOF {
					break
				}

				line = strings.TrimSpace(line)
				args := strings.Split(line, " ")
				switch args[0] {
				case "get":
					if len(args) <= 1 {
						continue
					}
					result, err = tydomClientAPI.Get(args[1])
				case "delete":
					if len(args) <= 1 {
						continue
					}
					result, err = tydomClientAPI.Delete(args[1])
				case "post":
					if len(args) <= 2 {
						continue
					}
					result, err = tydomClientAPI.Post(args[1], args[2])
				case "put":
					if len(args) <= 2 {
						continue
					}
					result, err = tydomClientAPI.Put(args[1], args[2])
				case "info":
					result, err = tydomClientAPI.GetInfo()
				case "access":
					result, err = tydomClientAPI.GetDevicesAccess()
				case "join_status":
					result, err = tydomClientAPI.GetDevicesInstallStatus()
				case "refresh":
					err = tydomClientAPI.RefreshAll()
				case "moments":
					result, err = tydomClientAPI.GetMomentsConfig()
				case "scenarios":
					result, err = tydomClientAPI.GetScenariosConfig()
				case "meta":
					result, err = tydomClientAPI.GetDeviceMetadata()
				case "configs":
					result, err = tydomClientAPI.GetUserConfig()
				case "cmeta":
					result, err = tydomClientAPI.GetDeviceConfigMetadata()
				case "areas":
					result, err = tydomClientAPI.GetAreas()
				case "data":
					result, err = tydomClientAPI.GetDevicesData()
				// case "password":
				case "quit":
					return
				default:
					fmt.Printf("Commands:\n info - get gateway infos\n password <new> <old> - change password\n")

					continue
				}
				if err != nil {
					slog.Error("error sending command", "error", err)

					return
				}
				if result != nil {
					prettyJSON, err := json.MarshalIndent(result, "", "  ")
					if err != nil {
						fmt.Println(result)
					} else {
						fmt.Println(string(prettyJSON))
					}
				}
			}
		},
	}

	return cliCmd
}
