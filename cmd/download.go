/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/rafaelbarbeta/MicroTorr/pkg/downloader"
	"github.com/rafaelbarbeta/MicroTorr/pkg/mtorr"
	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		verbosity, _ := cmd.Flags().GetInt("verbose")
		intNet, _ := cmd.Flags().GetString("interface")
		port, _ := cmd.Flags().GetString("port")
		if len(args) < 1 {
			fmt.Println("Error: You must specify a .mtorrent file")
			os.Exit(1)
		}
		mtorrent := mtorr.LoadMtorrent(args[0], verbosity)
		downloader.Download(mtorrent, intNet, port, verbosity)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	downloadCmd.Flags().IntP("verbose", "v", 0, "Choses verbosity level.")
	downloadCmd.Flags().StringP("interface", "i", "", "Specify the interface to retrieve IP from")
	downloadCmd.Flags().StringP("port", "p", "7777", "Specify the port to listen on for other peers in the swarm")
}
