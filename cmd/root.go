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

	// goro.yaml in the project root describes the project (ADR 0002), so commands run in the
	// root need no flag.
	rootCmd.PersistentFlags().StringVar(&goroCnf, "config", "goro.yaml", "path to the goro.yaml to use")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(updateCmd)
}
