/*
Copyright © 2024 Rafael Barbeta rafa.barbeta@gmail.com
*/
package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"

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
		seed, _ := cmd.Flags().GetString("seed")
		autoSeed, _ := cmd.Flags().GetBool("auto-seed")
		waitSeeders, _ := cmd.Flags().GetInt("waitSeeders")
		waitLeechers, _ := cmd.Flags().GetInt("waitLeechers")
		maxDownSpeed, _ := cmd.Flags().GetInt("max-down-speed")
		maxUpSpeed, _ := cmd.Flags().GetInt("max-up-speed")
		var err error
		if len(args) < 1 {
			fmt.Println("Error: You must specify a .mtorrent file")
			os.Exit(1)
		}
		if waitSeeders < 0 || waitLeechers < 0 {
			fmt.Println("Error: waitSeeders and waitLeechers must be greater than 0")
			os.Exit(1)
		}
		if seed != "" && (waitSeeders > 1 || waitLeechers > 0) {
			fmt.Println("Warning: waitSeeders and waitLeechers are ignored in seeding mode")
		}
		if intNet != "" {
			_, err = net.InterfaceByName(intNet)
		}
		if err != nil {
			fmt.Println("Error: Interface not found")
			os.Exit(1)
		}
		portInt, err := strconv.Atoi(port)
		if portInt >= 65535 || err != nil {
			fmt.Println("Error: Invalid port")
			os.Exit(1)
		}
		if maxDownSpeed < -1 || maxUpSpeed < -1 {
			fmt.Println("Error: max-down-speed and max-up-speed must be greater than -1")
		}
		mtorrent := mtorr.LoadMtorrent(args[0], verbosity)
		downloader.Download(mtorrent, intNet, port, seed, autoSeed, waitSeeders, waitLeechers, maxDownSpeed, maxUpSpeed, verbosity)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().IntP("verbose", "v", 0, "Choses verbosity level.")
	downloadCmd.Flags().StringP("interface", "i", "", "Specify the interface to retrieve IP from")
	downloadCmd.Flags().StringP("port", "p", "7777", "Specify the port to listen on for other peers in the swarm")
	downloadCmd.Flags().StringP("seed", "s", "", "Seed the torrent swarm with specified complete file")
	downloadCmd.Flags().BoolP("auto-seed", "a", false, "Wether to seed the file after download or not")
	downloadCmd.Flags().Int("waitSeeders", 1, "Number of seeders to wait for before download starts")
	downloadCmd.Flags().Int("waitLeechers", 0, "Number of leechers to wait for before download starts")
	downloadCmd.Flags().IntP("max-down-speed", "d", 0, "Specify the maximum download speed in KB/s. 0 for no limit")
	downloadCmd.Flags().IntP("max-up-speed", "u", 0, "Specify the maximum upload speed in KB/s. 0 for no limit")
}
