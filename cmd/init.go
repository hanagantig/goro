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
	Long: `Init creates a new service from goro.yaml: the one in the current directory, or the file
--config names. It asks for the app name and the work dir when goro.yaml leaves them empty; a
relative work_dir counts from the directory that holds goro.yaml.

Init saves a copy of goro.yaml in the work dir, the project root, with work_dir: . and the app
name if it asked for one, so that update and add find it there. When goro.yaml is in the work dir
already, init writes what it asked for into that file instead. It refuses a work dir that holds a
goro project already, or a different goro.yaml.`,
	Example: `  goro init --config ~/templates/goro.yaml`,
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
