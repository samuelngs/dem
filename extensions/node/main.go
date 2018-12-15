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
//     node:
//       version: 10.14.2

var nodeBinaryHost = "https://nodejs.org/dist"

type plugin struct {
	wsconf       *workspaceconfig.Config
	nodeconf     *nodeConfig
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
	Node *nodeConfig `yaml:"node"`
}

type nodeConfig struct {
	Version string `yaml:"version"`
}

func (v *plugin) Init(wsconf *workspaceconfig.Config) (bool, error) {
	var nodeconf *config
	if err := yaml.Unmarshal(wsconf.Src, &nodeconf); err != nil {
		return false, err
	}
	if nodeconf == nil || nodeconf.Workspace.With.Node == nil || len(nodeconf.Workspace.With.Node.Version) == 0 {
		return false, nil
	}
	v.wsconf = wsconf
	v.nodeconf = nodeconf.Workspace.With.Node
	v.refName = fmt.Sprintf("node-v%s-%s-x64", v.nodeconf.Version, runtime.GOOS)
	v.tarName = fmt.Sprintf("%s.tar.gz", v.refName)
	v.installURL = fmt.Sprintf("%s/v%s/%s", nodeBinaryHost, v.nodeconf.Version, v.tarName)
	v.installPath = filepath.Join(v.wsconf.InstallationDir, "node", v.nodeconf.Version)
	v.releasesPath = filepath.Join(v.wsconf.InstallationDir, "node", "releases")
	v.downloadPath = filepath.Join(v.releasesPath, v.tarName)
	v.binPath = filepath.Join(v.installPath, v.refName, "bin", "node")
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
	return composer.AsMap()
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
	if v.nodeconf != nil && len(v.nodeconf.Version) > 0 {
		return fmt.Sprintf("node %s", v.nodeconf.Version)
	}
	return "node"
}

// Export is a plugin instance used for workspace
var Export = ext.Extension(new(plugin))
