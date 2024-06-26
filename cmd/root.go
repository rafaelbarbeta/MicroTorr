/*
Copyright © 2024 Rafael Barbeta rafa.barbeta@gmail.com
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "MicroTorr",
	Short: "Simplistic implementation of BitTorrentV1.",
	Long: `Basic implementation of a peer-to-peer download client.

Inspired by the BitTorrent Protocol, for didatic purposes`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
