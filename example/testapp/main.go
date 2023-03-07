package main

import (
	"log"

	"github.com/spf13/cobra"

	"testapp/cmd"
	"testapp/internal/app"
)

// ldflags pass variables
var (
	commit      = "none"
	version     = "dev"
	serviceName = "testapp"
)

var configFilePath string

func initApp() {
	a, err := app.NewApp(configFilePath)
	if err != nil {
		log.Fatal("Fail to create app: ", err)
	}

	app.SetGlobalApp(a)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "ping-pong service",
		Short: "Main entry-point command for the application",
	}

	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "config file path")

	cobra.OnInitialize(initApp)

	rootCmd.AddCommand(
		cmd.RunHTTP(),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute root cmd: %v", err)

		return
	}
}
