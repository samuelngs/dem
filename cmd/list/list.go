// Package list implements the `list` command
package list

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/samuelngs/workspace/pkg/globalconfig"
	"github.com/samuelngs/workspace/pkg/workspaceconfig"
	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
	storageDir := os.ExpandEnv(globalconfig.Settings.StorageDir)

	files, err := ioutil.ReadDir(storageDir)
	if err != nil {
		return err
	}

	namespaces := make([]string, 0)
	for _, file := range files {
		workingDir := fmt.Sprintf("%s/%s", storageDir, file.Name())
		configPath := fmt.Sprintf("%s/%s", workingDir, ".workspace.yaml")
		if file.IsDir() && workspaceconfig.IsValid(configPath) {
			namespaces = append(namespaces, file.Name())
		}
	}

	fmt.Println(strings.Join(namespaces, "\n"))
	return nil
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all existing workspaces",
		Long:  "Lists all existing workspaces",
		RunE:  run,
	}
	return cmd
}
