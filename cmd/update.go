package cmd

import (
	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/pkg/log"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates your code after goro.yaml changes",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := commands.LoadConfig(goroCnf)
		if err != nil {
			log.Fatal(err)
		}

		_, err = commands.Update(cfg, commands.Options{})
		if err != nil {
			log.Fatal(err)
		}
	},
}
