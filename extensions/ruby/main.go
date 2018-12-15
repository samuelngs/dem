package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver"
	"github.com/samuelngs/dem/pkg/ext"
	"github.com/samuelngs/dem/pkg/util/downloader"
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
//     ruby:
//       version: 2.5.3

var rubyBinaryHost = "https://s3.amazonaws.com/travis-rubies/binaries"

type plugin struct {
	wsconf       *workspaceconfig.Config
	rubyconf     *rubyConfig
	binPath      string
	refName      string
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
	Ruby *rubyConfig `yaml:"ruby"`
}

type rubyConfig struct {
	Version string `yaml:"version"`
}

func (v *plugin) Init(wsconf *workspaceconfig.Config) (bool, error) {
	var rubyconf *config
	if err := yaml.Unmarshal(wsconf.Src, &rubyconf); err != nil {
		return false, err
	}
	if rubyconf == nil || rubyconf.Workspace.With.Ruby == nil || len(rubyconf.Workspace.With.Ruby.Version) == 0 {
		return false, nil
	}
	v.wsconf = wsconf
	v.rubyconf = rubyconf.Workspace.With.Ruby
	v.refName = fmt.Sprintf("ruby-%s", v.rubyconf.Version)
	v.tarName = fmt.Sprintf("%s.tar.bz2", v.refName)
	switch runtime.GOOS {
	case "darwin":
		v.installURL = fmt.Sprintf("%s/osx/10.13/x86_64/%s", rubyBinaryHost, v.tarName)
	case "linux":
		v.installURL = fmt.Sprintf("%s/ubuntu/16.04/x86_64/%s", rubyBinaryHost, v.tarName)
	default:
		return false, nil
	}
	v.installPath = filepath.Join(v.wsconf.InstallationDir, "ruby")
	v.releasesPath = filepath.Join(v.wsconf.InstallationDir, "ruby", "releases")
	v.downloadPath = filepath.Join(v.releasesPath, v.tarName)
	v.binPath = filepath.Join(v.installPath, v.refName, "bin", "ruby")
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
			if err := archiver.NewTarBz2().Unarchive(v.downloadPath, v.installPath); err != nil {
				return err
			}
			return nil
		}),
	}
}

func (v *plugin) Environment() map[string]string {
	return nil
}

func (v *plugin) Aliases() map[string]string {
	return nil
}

func (v *plugin) Sources() []string {
	return nil
}

func (v *plugin) Paths() []string {
	return []string{filepath.Join(v.installPath, v.refName, "bin")}
}

func (v *plugin) String() string {
	if v.rubyconf != nil && len(v.rubyconf.Version) > 0 {
		return fmt.Sprintf("ruby %s", v.rubyconf.Version)
	}
	return "ruby"
}

// Export is a plugin instance used for workspace
var Export = ext.Extension(new(plugin))
