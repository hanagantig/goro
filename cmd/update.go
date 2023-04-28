package cmd

import (
	"github.com/hanagantig/goro/internal/commands"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates your code after goro.yaml changes",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		commands.UpdateApp(goroCnf)
	},
}
