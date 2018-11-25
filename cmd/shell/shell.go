// Package shell implements the `shell` command
package shell

import (
	"fmt"
	"os"

	"github.com/samuelngs/workspace/pkg/globalconfig"
	"github.com/samuelngs/workspace/pkg/util/env"
	"github.com/samuelngs/workspace/pkg/util/exec"
	"github.com/samuelngs/workspace/pkg/util/fs"
	"github.com/samuelngs/workspace/pkg/util/homedir"
	"github.com/samuelngs/workspace/pkg/workspaceconfig"
	"github.com/spf13/cobra"
)

var key = "CWKS"

func shell(namespace string) error {
	storageDir := os.ExpandEnv(globalconfig.Settings.StorageDir)
	workingDir := fmt.Sprintf("%s/%s", storageDir, namespace)

	if !fs.Exists(workingDir) {
		fmt.Printf("workspace '%s' does not exist\n", namespace)
		return nil
	}

	configPath := fmt.Sprintf("%s/%s", workingDir, ".workspace.yaml")
	config, err := workspaceconfig.Load(configPath)
	if err != nil {
		fmt.Printf("(%s) unable to parse YAML configuration\n", namespace)
		return err
	}

	env := []string{
		// fixes issue where backspace behaves strangely with zsh
		fmt.Sprintf("TERM=%s", env.GetEnvAsString("TERM", "xterm")),
		fmt.Sprintf("SHELL=%s", config.Shell.Program),
		// maps virtual user to shell
		fmt.Sprintf("USER=%s", namespace),
		// maps virtual home to shell
		fmt.Sprintf("HOME=%s", "/"),
		// inject original home path to shell
		fmt.Sprintf("UNMASK_HOME=%s", homedir.Dir()),
		// patch shell prompt
		fmt.Sprintf("PS1=(%s) $ ", namespace),
		fmt.Sprintf("%s=1", key),
		// attempts to fix terminal copy and paste issue, it also
		// fixes X11 compatibility issue.
		fmt.Sprintf("DISPLAY=%s", env.GetEnvAsString("DISPLAY", ":0.0")),
	}
	for key, val := range config.Env {
		env = append(env, fmt.Sprintf("%s=%v", key, val))
	}

	cmd := exec.New(config.Shell.Program, config.Shell.Args...)

	cmd.SetDir(workingDir)
	cmd.SetEnv(env...)

	return cmd.Run()
}

func run(cmd *cobra.Command, args []string) error {
	if isInstance := env.Has(key); isInstance {
		return nil
	}
	if len(args) > 0 {
		return cmd.Usage()
	}
	shell(cmd.CalledAs())
	return nil
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand(namespace string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   fmt.Sprintf("%s [command]", namespace),
		Short:                 "Built-in magic commands",
		Long:                  "Built-in magic commands",
		DisableFlagsInUseLine: true,
		SilenceErrors:         true,
		Hidden:                true,
		RunE:                  run,
	}
	return cmd
}
