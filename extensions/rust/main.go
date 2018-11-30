package main

import (
	"bytes"
	"path/filepath"
	"time"

	"github.com/samuelngs/dem/pkg/ext"
	"github.com/samuelngs/dem/pkg/shell"
	"github.com/samuelngs/dem/pkg/util/envcomposer"
	"github.com/samuelngs/dem/pkg/util/fs"
	"github.com/samuelngs/dem/pkg/workspaceconfig"
	"gopkg.in/yaml.v2"
)

// Example of .workspace.yaml:
//
// workspace:
//   shell:
//     program: /bin/zsh
//   with:
//     rust:

type plugin struct {
	wsconf     *workspaceconfig.Config
	cargoPath  string
	rustupPath string
}

type config struct {
	Workspace *workspaceConfig `yaml:"workspace"`
}

type workspaceConfig struct {
	With *withConfig `yaml:"with"`
}

type withConfig struct {
	Rust *rustConfig `yaml:"rust"`
}

type rustConfig struct {
}

func (v *plugin) Init(wsconf *workspaceconfig.Config) (bool, error) {
	var rustconf *config
	if err := yaml.Unmarshal(wsconf.Src, &rustconf); err != nil {
		return false, err
	}
	v.wsconf = wsconf
	v.cargoPath = filepath.Join(v.wsconf.InstallationDir, ".cargo")
	v.rustupPath = filepath.Join(v.wsconf.InstallationDir, ".multirust")
	return rustconf != nil, nil
}

func (v *plugin) SetupTasks() ext.SetupTasks {
	config := v.wsconf
	if fs.Exists(v.cargoPath) && fs.Exists(v.rustupPath) {
		return nil
	}
	return ext.SetupTasks{
		ext.Procedure("installing", func(bar ext.ProgressBar) error {
			go func() {
				for !bar.Completed() {
					bar.IncrBy(1)
					time.Sleep(2 * time.Second)
				}
			}()
			silent := new(bytes.Buffer)
			envcomposer := envcomposer.New()
			envcomposer.Set("CARGO_HOME", v.cargoPath)
			envcomposer.Set("RUSTUP_HOME", v.rustupPath)
			envcomposer.Set("SHELL", config.Workspace.Shell.Program)
			envcomposer.Set("USER", config.Namespace)
			envcomposer.Set("HOME", config.WorkingDir)
			cmd := shell.New(v.wsconf.Workspace.Shell.Program, "-c", "curl https://sh.rustup.rs -sSf | sh -s -- -y --no-modify-path")
			cmd.SetDir(config.WorkingDir)
			cmd.SetEnv(envcomposer.AsMap())
			cmd.SetStdout(silent)
			cmd.SetStdin(silent)
			cmd.SetStderr(silent)
			return cmd.Run()
		}),
	}
}

func (v *plugin) Environment() map[string]string {
	return map[string]string{
		"CARGO_HOME":  v.cargoPath,
		"RUSTUP_HOME": v.rustupPath,
	}
}

func (v *plugin) Aliases() map[string]string {
	return nil
}

func (v *plugin) Sources() []string {
	return nil
}

func (v *plugin) Paths() []string {
	return []string{filepath.Join(v.cargoPath, "bin")}
}

func (v *plugin) String() string {
	return "rust"
}

// Export is a plugin instance used for workspace
var Export = ext.Extension(new(plugin))
