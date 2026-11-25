package cmd

import (
	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/pkg/log"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new service",
	Long:  `Create service based on goro.yaml config or add needed parameters manually`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := commands.LoadConfig(goroCnf)
		if err != nil {
			log.Fatal(err)
		}

		err = commands.Init(cfg, commands.Options{})
		if err != nil {
			log.Fatal(err)
		}
	},
}
