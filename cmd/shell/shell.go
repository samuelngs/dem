// Package shell implements the `shell` command
package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/samuelngs/workspace/pkg/ext"
	"github.com/samuelngs/workspace/pkg/globalconfig"
	"github.com/samuelngs/workspace/pkg/util/env"
	"github.com/samuelngs/workspace/pkg/util/envcomposer"
	"github.com/samuelngs/workspace/pkg/util/exec"
	"github.com/samuelngs/workspace/pkg/util/fs"
	"github.com/samuelngs/workspace/pkg/util/homedir"
	"github.com/samuelngs/workspace/pkg/workspaceconfig"
	"github.com/spf13/cobra"
)

var key = "CWKS"

func extensions(config *workspaceconfig.Config) []ext.Extension {
	extensions := make([]ext.Extension, 0)
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
	return extensions
}

func shell(namespace string) error {
	storageDir := os.ExpandEnv(globalconfig.Settings.StorageDir)
	pluginsDir := os.ExpandEnv(globalconfig.Settings.PluginsDir)
	workingDir := fmt.Sprintf("%s/%s", storageDir, namespace)

	if !fs.Exists(workingDir) {
		fmt.Printf("workspace '%s' does not exist\n", namespace)
		return nil
	}

	configPath := fmt.Sprintf("%s/%s", workingDir, ".workspace.yaml")

	yaml, err := workspaceconfig.Read(configPath)
	if err != nil {
		fmt.Printf("(%s) unable to read YAML configuration\n", namespace)
		return err
	}

	config, err := workspaceconfig.Parse(yaml)
	if err != nil {
		fmt.Printf("(%s) unable to parse YAML configuration\n", namespace)
		return err
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
	paths := make([]string, 0)
	for _, ext := range exts {
		for key, val := range ext.Environment() {
			envcomposer.Set(key, val)
		}
		paths = append(paths, ext.Bin()...)
	}

	// TODO OOOOOOOOOOOO BLAH BLAH BLAHHHHHHH
	// path := fmt.Sprintf("PATH=%s:$PATH", strings.Join(paths, ":"))

	cmd := exec.New(config.Workspace.Shell.Program, config.Workspace.Shell.Args...)

	cmd.SetDir(workingDir)
	cmd.SetEnv(envcomposer.AsArray()...)

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
	if err := shell(cmd.CalledAs()); err != nil {
		fmt.Print(err)
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
