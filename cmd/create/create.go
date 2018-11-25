// Package create implements the `create` command
package create

import (
	"fmt"
	"os"
	"strings"

	"github.com/samuelngs/workspace/pkg/globalconfig"
	"github.com/samuelngs/workspace/pkg/util/fs"
	"github.com/samuelngs/workspace/pkg/workspaceconfig"
	"github.com/spf13/cobra"
)

func createWorkspace(namespace string) error {
	storageDir := os.ExpandEnv(globalconfig.Settings.StorageDir)
	workingDir := fmt.Sprintf("%s/%s", storageDir, namespace)
	configPath := fmt.Sprintf("%s/%s", workingDir, ".workspace.yaml")

	// cancel action if workspace already exists
	if fs.Exists(configPath) {
		fmt.Printf("workspace '%s' already exists\n", namespace)
		return nil
	}

	// initialize workspace directory
	if err := fs.Mkdir(workingDir); err != nil {
		return err
	}

	// initialize default workspace settings
	b, err := workspaceconfig.New()
	if err != nil {
		return err
	}

	if err := fs.WriteFile(configPath, b); err != nil {
		return err
	}

	fmt.Printf("workspace '%s' created\n", namespace)
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	switch {
	case len(args) == 0:
		return cmd.Usage()
	case len(strings.TrimSpace(args[0])) == 0:
		return cmd.Usage()
	default:
		return createWorkspace(args[0])
	}
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "create [namespace]",
		Short:                 "Creates an isolated development workspace [name]",
		Long:                  "Creates a local isolated development workspace",
		Aliases:               []string{"init", "add", "up"},
		DisableFlagsInUseLine: true,
		RunE:                  run,
	}
	return cmd
}
