// Package delete implements the `delete` command
package delete

import (
	"fmt"
	"os"
	"strings"

	"github.com/samuelngs/workspace/pkg/globalconfig"
	"github.com/samuelngs/workspace/pkg/util/fs"
	"github.com/spf13/cobra"
)

var keepFiles bool

func deleteWorkspace(namespace string) error {
	storageDir := os.ExpandEnv(globalconfig.Settings.StorageDir)
	workingDir := fmt.Sprintf("%s/%s", storageDir, namespace)
	configPath := fmt.Sprintf("%s/%s", workingDir, ".workspace.yaml")

	var path string
	if keepFiles {
		path = configPath
	} else {
		path = workingDir
	}

	if !fs.Exists(path) {
		fmt.Printf("workspace '%s' does not exist\n", namespace)
		return nil
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	fmt.Printf("workspace '%s' deleted\n", namespace)
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	switch {
	case len(args) == 0:
		return cmd.Usage()
	case len(strings.TrimSpace(args[0])) == 0:
		return cmd.Usage()
	default:
		return deleteWorkspace(args[0])
	}
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes one or more workspaces",
		Long:  "Deletes one or more workspaces",
		RunE:  run,
	}
	cmd.PersistentFlags().BoolVarP(&keepFiles, "keep-files", "k", true, "keep workspace files")
	return cmd
}
