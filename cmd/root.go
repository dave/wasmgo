package cmd

import (
	"fmt"
	"os"

	"github.com/dave/wasmgo/cmd/cmdconfig"
	"github.com/spf13/cobra"
)

var global = &cmdconfig.Config{}

var rootCmd = &cobra.Command{
	Use:   "wasmgo",
	Short: "Compile Go to WASM, test locally or deploy to jsgo.io",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
