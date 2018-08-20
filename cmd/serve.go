package cmd

import (
	"fmt"
	"os"

	"github.com/dave/wasmgo/cmd/server"
	"github.com/spf13/cobra"
)

func init() {
	serveCmd.PersistentFlags().IntVarP(&global.Port, "port", "p", 8080, "Server port.")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve [package]",
	Short: "Serve locally",
	Long:  "Starts a webserver locally, and recompiles the WASM on every page refresh, for testing and development.",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			global.Path = args[0]
		}
		if err := server.Start(global); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}
	},
}
