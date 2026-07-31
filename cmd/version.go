/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/dchauviere/go-tydom/internal/config"
	"github.com/spf13/cobra"
)

type versionCmd struct {
	showBuildInfo bool
}

func (vc *versionCmd) Command() *cobra.Command {
	// versionCmd represents the version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			if debugInfo, ok := debug.ReadBuildInfo(); ok && vc.showBuildInfo {
				fmt.Printf("GO-Tydom\nVersion %s\nCommit %s\n\n%v\n", config.VERSION, config.COMMIT, debugInfo)
			} else {
				fmt.Printf("GO-Tydom\nVersion %s\nCommit %s\n", config.VERSION, config.COMMIT)
			}
		},
	}
	versionCmd.Flags().BoolVarP(&vc.showBuildInfo, "verbose", "v", false, "Show build info")
	return versionCmd
}
