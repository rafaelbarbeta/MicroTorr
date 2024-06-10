/*
Copyright © 2024 Rafael Barbeta rafa.barbeta@gmail.com
*/
package cmd

import (
	"log"
	"net/http"

	//"encoding/json"
	"github.com/mitchellh/colorstring"
	"github.com/rafaelbarbeta/MicroTorr/pkg/tracker"
	"github.com/spf13/cobra"
)

// trackerCmd represents the tracker command
var trackerCmd = &cobra.Command{
	Use:   "tracker",
	Short: "Start a HTTP server to act as a tracker",
	Long: `HTTP server that will be used as a tracker for mtorrent clients.
	
Once it is activated, it will bind to port 8888 by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		bind, _ := cmd.Flags().GetString("bind")
		verbosity, _ := cmd.Flags().GetInt("verbosity")
		colorstring.Println("Tracker serving on: " + "[red]" + bind)
		http.HandleFunc("/announce", func(w http.ResponseWriter, r *http.Request) {
			tracker.Announce(w, r, verbosity)
		})
		log.Fatal(http.ListenAndServe(bind, nil))
	},
}

func init() {
	rootCmd.AddCommand(trackerCmd)

	trackerCmd.Flags().StringP("bind", "b", "0.0.0.0:8888", "Specify the address to bind")
	trackerCmd.Flags().IntP("verbosity", "v", 0, "Choses verbosity level.")
}
