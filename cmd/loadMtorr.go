/*
Copyright © 2024 Rafael Barbeta rafa.barbeta@gmail.com
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
		fmt.Println(mtorrent)
	},
}

func init() {
	rootCmd.AddCommand(loadMtorrCmd)
}
