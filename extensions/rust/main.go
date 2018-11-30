package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samuelngs/dem/pkg/ext"
	"github.com/samuelngs/dem/pkg/shell"
	"github.com/samuelngs/dem/pkg/util/downloader"
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
//       version: 1.29.2

type plugin struct {
	wsconf            *workspaceconfig.Config
	rsconf            *rustConfig
	cargoPath         string
	rustupPath        string
	utlityPath        string
	installScriptPath string
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
	Version string `yaml:"version"`
}

func (v *plugin) Init(wsconf *workspaceconfig.Config) (bool, error) {
	var rustconf *config
	if err := yaml.Unmarshal(wsconf.Src, &rustconf); err != nil {
		return false, err
	}
	if rustconf == nil || rustconf.Workspace.With.Rust == nil || len(rustconf.Workspace.With.Rust.Version) == 0 {
		return false, nil
	}
	v.wsconf = wsconf
	v.rsconf = rustconf.Workspace.With.Rust
	v.cargoPath = filepath.Join(v.wsconf.InstallationDir, "rust", v.rsconf.Version, ".cargo")
	v.rustupPath = filepath.Join(v.wsconf.InstallationDir, "rust", v.rsconf.Version, ".multirust")
	v.utlityPath = filepath.Join(v.wsconf.InstallationDir, "rust", "helper")
	v.installScriptPath = filepath.Join(v.utlityPath, "rustup")
	return true, nil
}

func (v *plugin) SetupTasks() ext.SetupTasks {
	config := v.wsconf
	if fs.Exists(v.cargoPath) && fs.Exists(v.rustupPath) {
		return nil
	}
	return ext.SetupTasks{
		ext.Procedure("initializing", func(bar ext.ProgressBar) error {
			return fs.Mkdir(v.utlityPath)
		}),
		ext.Procedure("downloading", func(bar ext.ProgressBar) error {
			cb := make(chan int)
			go func() {
				var lp int
				for progress := range cb {
					if !bar.Completed() {
						bar.IncrBy(progress - lp)
						lp = progress
					}
				}
			}()
			return downloader.New("https://sh.rustup.rs", v.installScriptPath).Start(cb)
		}),
		ext.Procedure("installing", func(bar ext.ProgressBar) error {
			if err := os.Chmod(v.installScriptPath, 0755); err != nil {
				return err
			}
			envcomposer := envcomposer.New()
			envcomposer.Set("CARGO_HOME", v.cargoPath)
			envcomposer.Set("RUSTUP_HOME", v.rustupPath)
			envcomposer.Set("SHELL", config.Workspace.Shell.Program)
			envcomposer.Set("USER", config.Namespace)
			envcomposer.Set("HOME", config.WorkingDir)
			cmd := shell.New(v.installScriptPath, "--no-modify-path", "--default-toolchain", v.rsconf.Version, "-y")
			cmd.SetDir(config.WorkingDir)
			cmd.SetEnv(envcomposer.AsMap())
			cmd.SetStdin(nil)
			cmd.SetStdout(nil)
			cmd.SetStderr(nil)
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
	if v.rsconf != nil && len(v.rsconf.Version) > 0 {
		return fmt.Sprintf("rust %s", v.rsconf.Version)
	}
	return "rust"
}

// Export is a plugin instance used for workspace
var Export = ext.Extension(new(plugin))
