/*
Copyright © 2024 Rafael Barbeta rafa.barbeta@gmail.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/rafaelbarbeta/MicroTorr/pkg/mtorr"

	"github.com/spf13/cobra"
)

// createMtorrCmd represents the createMtorr command
var createMtorrCmd = &cobra.Command{
	Use:   "createMtorr",
	Short: "Generate a .mtorrent for a file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		tracker, _ := cmd.Flags().GetString("tracker")
		pieceLength, _ := cmd.Flags().GetInt("pieceLength")
		verbose, _ := cmd.Flags().GetInt("verbose")
		if len(args) < 1 {
			fmt.Println("Error: You need to specify a file to create torrent from")
			os.Exit(1)
		}
		mtorr.GenMtorrent(args[0], tracker, pieceLength, verbose)
	},
}

func init() {
	rootCmd.AddCommand(createMtorrCmd)
	createMtorrCmd.Flags().StringP("tracker", "t", "http://127.0.0.1:8888", "Specify a URL tracker for this file")
	createMtorrCmd.Flags().IntP("pieceLength", "l", 16000, "Specify the length of each piece. Default: 16KB")
	createMtorrCmd.Flags().IntP("verbose", "v", 0, "Choses verbosity level.")
}
