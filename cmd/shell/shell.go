// Package shell implements the `shell` command
package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/samuelngs/workspace/pkg/ext"
	"github.com/samuelngs/workspace/pkg/globalconfig"
	"github.com/samuelngs/workspace/pkg/log"
	"github.com/samuelngs/workspace/pkg/shell"
	"github.com/samuelngs/workspace/pkg/util/env"
	"github.com/samuelngs/workspace/pkg/util/envcomposer"
	"github.com/samuelngs/workspace/pkg/util/fs"
	"github.com/samuelngs/workspace/pkg/util/homedir"
	"github.com/samuelngs/workspace/pkg/workspaceconfig"
	"github.com/spf13/cobra"
)

var key = "CWKS"

func extensions(config *workspaceconfig.Config) []ext.Extension {
	extensions := make([]ext.Extension, 0)
	if config.Workspace.With != nil {
		modules, err := filepath.Glob(fmt.Sprintf("%s/*.so", config.PluginsDir))
		if err != nil {
			return nil
		}
		for _, module := range modules {
			p, err := plugin.Open(module)
			if err != nil {
				continue
			}
			v, err := p.Lookup("Export")
			if err != nil {
				continue
			}
			i, ok := v.(*ext.Extension)
			if !ok {
				continue
			}
			m := *i
			success, err := m.Init(config)
			if !success || err != nil {
				continue
			}
			extensions = append(extensions, m)
		}
	}
	return extensions
}

func createSession(namespace string) error {
	storageDir := os.ExpandEnv(globalconfig.Settings.StorageDir)
	pluginsDir := os.ExpandEnv(globalconfig.Settings.PluginsDir)
	workingDir := fmt.Sprintf("%s/%s", storageDir, namespace)

	if !fs.Exists(workingDir) {
		return fmt.Errorf("workspace '%s' does not exist", namespace)
	}

	configPath := fmt.Sprintf("%s/%s", workingDir, ".workspace.yaml")

	yaml, err := workspaceconfig.Read(configPath)
	if err != nil {
		return fmt.Errorf("(%s) unable to read YAML configuration", namespace)
	}

	config, err := workspaceconfig.Parse(yaml)
	if err != nil {
		return fmt.Errorf("(%s) unable to parse YAML configuration", namespace)
	}
	config.Namespace = namespace
	config.WorkingDir = workingDir
	config.PluginsDir = pluginsDir
	config.Src = yaml

	// load workspace extensions
	exts := extensions(config)

	// environment composer
	envcomposer := envcomposer.New()

	// fixes issue where backspace behaves strangely with zsh
	envcomposer.Set("TERM", env.GetEnvAsString("TERM", "xterm"))
	envcomposer.Set("SHELL", config.Workspace.Shell.Program)
	// maps virtual user to shell
	envcomposer.Set("USER", namespace)
	envcomposer.Set("HOME", workingDir)
	envcomposer.Set("UNMASK_HOME", homedir.Dir())
	envcomposer.Set("PS1", fmt.Sprintf("(%s) $ ", namespace))
	envcomposer.Set(key, "1")
	// attempts to fix terminal copy and paste issue, it also
	// fixes X11 compatibility issue.
	envcomposer.Set("DISPLAY", env.GetEnvAsString("DISPLAY", ":0.0"))

	for key, val := range config.Workspace.Environment {
		envcomposer.Set(key, val)
	}

	// prepare extensions environment variables and bin paths
	var (
		paths   = make([]string, 0)
		aliases = config.Workspace.Aliases
	)
	for _, ext := range exts {
		for key, val := range ext.Environment() {
			envcomposer.Set(key, val)
		}
		for alias, cmd := range ext.Aliases() {
			aliases[alias] = cmd
		}
		paths = append(paths, ext.Bin()...)
	}
	envcomposer.Set("EXT_PATH", strings.Join(paths, ":"))

	cmd := shell.New(config.Workspace.Shell.Program, config.Workspace.Shell.Args...)

	cmd.SetDir(workingDir)
	cmd.SetEnv(envcomposer.AsMap())
	cmd.SetAliases(aliases)

	for _, ext := range exts {
		if err := ext.StartPre(); err != nil {
			return err
		}
	}

	return cmd.Run()
}

func run(cmd *cobra.Command, args []string) error {
	if isInstance := env.Has(key); isInstance {
		return nil
	}
	if len(args) > 0 {
		return cmd.Usage()
	}
	if err := createSession(cmd.CalledAs()); err != nil {
		log.Debug(err)
	}
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
