package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/pkg/log"
)

var (
	addDryRun bool

	addServiceDeps    []string
	addServiceMethods []string

	addAdapterStorage string
	addAdapterMethods []string
)

// addCmd groups the goro add subcommands (ADR 0002).
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a use case method, a service, a method or an adapter",
	Long: `Add records one new part of the service in goro.yaml and then runs update, so the code
follows in one step.

It reads goro.yaml in the current directory, or the file --config names, and never asks for
anything. It only inserts lines into goro.yaml: comments, blank lines and the order of keys stay
as they are, and a list written as [a, b] is refused because it can't grow by a line. A change
goro.yaml can't take, such as a name it lists already or a dependency it doesn't list, is an error
and changes nothing.

Use --dry-run to see what add would do, with the goro.yaml diff, and write nothing.`,
	Example: `  goro add usecase SignOut
  goro add service Billing --deps PaymentsRepo --methods Charge,Refund
  goro add method Billing.Void
  goro add adapter PaymentsRepo --storage pgsqlx --methods Save`,
}

var addUseCaseCmd = &cobra.Command{
	Use:     "usecase <Method>",
	Short:   "Adds a use case method",
	Long:    `Adds Method to use_case.methods in goro.yaml, which creates its file in internal/usecase.`,
	Example: `  goro add usecase SignOut`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runAdd(commands.AddUseCase(goroCnf, args[0], addOptions()))
	},
}

var addServiceCmd = &cobra.Command{
	Use:   "service <Name>",
	Short: "Adds a service",
	Long: `Adds the service Name to services in goro.yaml, with its methods and dependencies: update
creates its package with a file per method and a dep_*.go file per dependency, and wires it in
internal/app/container.go. Each dependency is a service or an adapter goro.yaml lists already.

To add a dependency to a service that exists, edit goro.yaml and run goro update.`,
	Example: `  goro add service Billing --deps PaymentsRepo --methods Charge,Refund
  goro add service Mailer`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		svc := config.Service{Name: args[0], Deps: addServiceDeps, Methods: addServiceMethods}
		runAdd(commands.AddService(goroCnf, svc, addOptions()))
	},
}

var addMethodCmd = &cobra.Command{
	Use:   "method <Name>.<Method>",
	Short: "Adds a method to a service or an adapter",
	Long: `Adds Method to the methods of Name in goro.yaml, which creates its file in the package of Name.
Name is a service or an adapter goro.yaml lists; it fails when Name is both, or neither.`,
	Example: `  goro add method Billing.Void
  goro add method PaymentsRepo.Save`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runAdd(commands.AddMethod(goroCnf, args[0], addOptions()))
	},
}

var addAdapterCmd = &cobra.Command{
	Use:   "adapter <Name> --storage <storage>",
	Short: "Adds an adapter",
	Long: `Adds the adapter Name to adapters in goro.yaml, with its storage and methods, and adds the
storage to storages when goro.yaml doesn't list it yet: update creates the adapter package with a
file per method, and the storage's connection when it is new.`,
	Example: `  goro add adapter PaymentsRepo --storage pgsqlx --methods Save
  goro add adapter AuditAPI --storage http`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		adapter := config.Adapter{Name: args[0], Storage: config.Storage(addAdapterStorage), Methods: addAdapterMethods}
		runAdd(commands.AddAdapter(goroCnf, adapter, addOptions()))
	},
}

func addOptions() commands.Options {
	return commands.Options{DryRun: addDryRun}
}

func runAdd(_ commands.Report, err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	addCmd.PersistentFlags().BoolVar(&addDryRun, "dry-run", false,
		"print what add would do, with diffs of goro.yaml and of each regenerated file, and write nothing")

	addServiceCmd.Flags().StringSliceVar(&addServiceDeps, "deps", nil, "services and adapters the service depends on, comma-separated")
	addServiceCmd.Flags().StringSliceVar(&addServiceMethods, "methods", nil, "methods of the service, comma-separated")

	addAdapterCmd.Flags().StringVar(&addAdapterStorage, "storage", "",
		"the storage the adapter uses: one of "+strings.Join(config.SupportedStorages(), ", "))
	addAdapterCmd.Flags().StringSliceVar(&addAdapterMethods, "methods", nil, "methods of the adapter, comma-separated")
	_ = addAdapterCmd.MarkFlagRequired("storage")

	addCmd.AddCommand(addUseCaseCmd, addServiceCmd, addMethodCmd, addAdapterCmd)
}
