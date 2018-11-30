package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver"
	"github.com/samuelngs/dem/pkg/ext"
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
//     go:
//       version: 1.11.2
//       go_path: false
//       go_111_module: auto

var goBinaryHost = "https://dl.google.com/go"

type plugin struct {
	wsconf       *workspaceconfig.Config
	goconf       *goConfig
	binPath      string
	tarName      string
	releasesPath string
	installURL   string
	installPath  string
	downloadPath string
}

type config struct {
	Workspace *workspaceConfig `yaml:"workspace"`
}

type workspaceConfig struct {
	With *withConfig `yaml:"with"`
}

type withConfig struct {
	Go *goConfig `yaml:"go"`
}

type goConfig struct {
	Version     string `yaml:"version"`
	GoPath      string `yaml:"go_path"`
	Go111Module string `yaml:"go_111_module"`
}

func (v *plugin) Init(wsconf *workspaceconfig.Config) (bool, error) {
	var goconf *config
	if err := yaml.Unmarshal(wsconf.Src, &goconf); err != nil {
		return false, err
	}
	if goconf == nil || goconf.Workspace.With.Go == nil || len(goconf.Workspace.With.Go.Version) == 0 {
		return false, nil
	}
	v.wsconf = wsconf
	v.goconf = goconf.Workspace.With.Go
	v.tarName = fmt.Sprintf("go%s.%s-%s.tar.gz", v.goconf.Version, runtime.GOOS, runtime.GOARCH)
	v.installURL = fmt.Sprintf("%s/%s", goBinaryHost, v.tarName)
	v.installPath = filepath.Join(v.wsconf.InstallationDir, "go", v.goconf.Version)
	v.releasesPath = filepath.Join(v.wsconf.InstallationDir, "go", "releases")
	v.downloadPath = filepath.Join(v.releasesPath, v.tarName)
	v.binPath = filepath.Join(v.installPath, "go", "bin", "go")
	return true, nil
}

func (v *plugin) SetupTasks() ext.SetupTasks {
	if fs.Exists(v.binPath) {
		return nil
	}
	return ext.SetupTasks{
		ext.Procedure("initializing", func(bar ext.ProgressBar) error {
			return fs.Mkdir(v.installPath, v.releasesPath)
		}),
		ext.Procedure("downloading", func(bar ext.ProgressBar) error {
			cb := make(chan int)
			go func() {
				var lp int
				for progress := range cb {
					bar.IncrBy(progress - lp)
					lp = progress
				}
			}()
			return downloader.New(v.installURL, v.downloadPath).Start(cb)
		}),
		ext.Procedure("unpacking", func(bar ext.ProgressBar) error {
			if err := archiver.NewTarGz().Unarchive(v.downloadPath, v.installPath); err != nil {
				return err
			}
			return nil
		}),
	}
}

func (v *plugin) Environment() map[string]string {
	composer := envcomposer.New()
	if len(v.goconf.GoPath) > 0 && v.goconf.GoPath != "false" {
		composer.Set("GOPATH", v.goconf.GoPath)
	}
	if len(v.goconf.Go111Module) > 0 {
		composer.Set("GO111MODULE", v.goconf.Go111Module)
	}
	return composer.AsMap()
}

func (v *plugin) Aliases() map[string]string {
	return nil
}

func (v *plugin) Sources() []string {
	return nil
}

func (v *plugin) Paths() []string {
	return []string{filepath.Join(v.installPath, "go", "bin")}
}

func (v *plugin) String() string {
	if v.goconf != nil && len(v.goconf.Version) > 0 {
		return fmt.Sprintf("go %s", v.goconf.Version)
	}
	return "go"
}

// Export is a plugin instance used for workspace
var Export = ext.Extension(new(plugin))
