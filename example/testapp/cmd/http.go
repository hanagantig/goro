package cmd

import (
	"github.com/spf13/cobra"
	"testapp/internal/app"
)

func RunHTTP() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Run http server",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		a, err := app.GetGlobalApp()
		if err != nil {
			return err
		}

		if err := a.StartHTTPServer(); err != nil {
			return err
		}

		return nil
	}

	return cmd
}
