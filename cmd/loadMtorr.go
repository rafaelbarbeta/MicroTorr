/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/rafaelbarbeta/MicroTorr/pkg/mtorr"
	"github.com/spf13/cobra"
)

// loadMtorrCmd represents the loadMtorr command
var loadMtorrCmd = &cobra.Command{
	Use:   "loadMtorr",
	Short: "Load a .mtorrent file",
	Long:  `Load a .mtorrent file, and show its information.`,
	Run: func(cmd *cobra.Command, args []string) {
		mtorrent := mtorr.LoadMtorrent(args[0])
		fmt.Println("Tracker Link:", mtorrent.Announce)
		fmt.Println("File Name:", mtorrent.Info.Name)
		fmt.Println("File Length:", mtorrent.Info.Length)
		fmt.Println("Piece Length:", mtorrent.Info.Piece_length)
		fmt.Println("Sha1sum (first 20 bytes):", mtorrent.Info.Sha1sum[:20])
		fmt.Println("Id Hash:", mtorrent.Info.Id)
	},
}

func init() {
	rootCmd.AddCommand(loadMtorrCmd)
}
