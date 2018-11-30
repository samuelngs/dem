package main

import (
	"os"
	"strings"

	"github.com/samuelngs/dem/cmd/create"
	"github.com/samuelngs/dem/cmd/delete"
	"github.com/samuelngs/dem/cmd/describe"
	"github.com/samuelngs/dem/cmd/list"
	"github.com/samuelngs/dem/cmd/shell"
	"github.com/samuelngs/dem/pkg/globalconfig"
	"github.com/samuelngs/dem/pkg/util/fs"
	"github.com/samuelngs/dem/pkg/util/homedir"
	"github.com/spf13/cobra"
)

var conf string

func pre(cmd *cobra.Command, args []string) {
	if err := globalconfig.Load(conf); err != nil {
		os.Exit(1)
	}
	fs.Mkdir(os.ExpandEnv(globalconfig.Settings.StorageDir))
	fs.Mkdir(os.ExpandEnv(globalconfig.Settings.PluginsDir))
}

func run(cmd *cobra.Command, args []string) error {
	switch {
	case len(args) == 0:
		return cmd.Help()
	case len(strings.TrimSpace(args[0])) == 0:
		return cmd.Help()
	default:
		return nil
	}
}

// NewCommand returns a new cobra.Command implementing the root command for kind
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "dem [namespace] [command]",
		Short:                 "dem is a tool for managing isolated development workspaces",
		Long:                  "dem creates and manages isolated development workspaces",
		DisableFlagsInUseLine: true,
		SilenceErrors:         true,
		PersistentPreRun:      pre,
		RunE:                  run,
	}

	// hide help command in commands list
	cmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	// workspace available comments
	cmd.AddCommand(create.NewCommand())
	cmd.AddCommand(delete.NewCommand())
	cmd.AddCommand(describe.NewCommand())
	cmd.AddCommand(list.NewCommand())

	// too magical for this crap
	if args := os.Args[1:]; len(args) > 0 {
		cmd.AddCommand(shell.NewCommand(args[0]))
	}

	// command line flags
	cmd.PersistentFlags().StringVarP(&conf, "config", "c", homedir.Path(".dem.yaml"), "Location of config file")

	return cmd
}

// Run runs the `workspace` root command
func Run() error {
	return NewCommand().Execute()
}

func main() {
	if err := Run(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(-1)
	}
}
