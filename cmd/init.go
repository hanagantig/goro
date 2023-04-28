package cmd

import (
	"github.com/hanagantig/goro/internal/commands"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new service",
	Long:  `Create service based on goro.yaml config or add needed parameters manually`,
	Run: func(cmd *cobra.Command, args []string) {
		commands.InitApp(goroCnf)
	},
}
