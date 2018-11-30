package edit

import (
	"fmt"
	"os"

	"github.com/samuelngs/dem/pkg/globalconfig"
	"github.com/samuelngs/dem/pkg/shell"
	"github.com/samuelngs/dem/pkg/util/env"
	"github.com/spf13/cobra"
)

var namespace string

func run(cmd *cobra.Command, args []string) error {
	storageDir := os.ExpandEnv(globalconfig.Settings.StorageDir)
	workingDir := fmt.Sprintf("%s/%s", storageDir, namespace)
	configPath := fmt.Sprintf("%s/%s", workingDir, ".workspace.yaml")

	c := shell.New(env.GetEnvAsString("EDITOR", "vi"), configPath)
	c.SetDir(workingDir)

	return c.Run()
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand(ns string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "edit",
		Aliases:               []string{"update"},
		DisableFlagsInUseLine: true,
		RunE:                  run,
	}
	namespace = ns
	return cmd
}
