package cmd

import (
	"fmt"
	"os"

	"github.com/dave/wasmgo/cmd/cmdconfig"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&global.Index, "index", "i", "index.wasmgo.html", "Specify the index page template. Variables: Script, Loader, Binary.")
	rootCmd.PersistentFlags().BoolVarP(&global.Verbose, "verbose", "v", false, "Show detailed status messages.")
	rootCmd.PersistentFlags().BoolVarP(&global.Open, "open", "o", true, "Open the page in a browser (default true).")
	rootCmd.PersistentFlags().StringVarP(&global.Command, "command", "c", "go", "Name of the go command.")
	rootCmd.PersistentFlags().StringVarP(&global.Flags, "flags", "f", "", "Flags to pass to the go build command.")
	rootCmd.PersistentFlags().StringVarP(&global.BuildTags, "build", "b", "", "Build tags to pass to the go build command.")
}

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
