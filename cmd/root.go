package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var goroCnf string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goro",
	Short: "Goro app",
	Long:  ``,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVar(&goroCnf, "config", "", "path to goro yaml file")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(updateCmd)
}
