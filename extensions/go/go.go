package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver"
	"github.com/samuelngs/workspace/pkg/ext"
	"github.com/samuelngs/workspace/pkg/log"
	"github.com/samuelngs/workspace/pkg/util/envcomposer"
	"github.com/samuelngs/workspace/pkg/util/fs"
	"github.com/samuelngs/workspace/pkg/workspaceconfig"
	"github.com/sirupsen/logrus"
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
//       goPath: false
//       go111Module: auto

var goBinaryHost = "https://dl.google.com/go"

type plugin struct {
	wsconf *workspaceconfig.Config
	goconf *goConfig
	status *log.Status
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
	GoPath      string `yaml:"goPath"`
	Go111Module string `yaml:"go111Module"`
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
	v.status = log.NewStatus(os.Stdout)
	v.status.MaybeWrapLogrus(logrus.StandardLogger())
	return true, nil
}

func (v *plugin) StartPre() error {
	var (
		path = filepath.Join(v.wsconf.WorkingDir, ".go", v.goconf.Version)
		bin  = filepath.Join(path, "go", "bin", "go")
		tar  = fmt.Sprintf("go%s.%s-%s.tar.gz", v.goconf.Version, runtime.GOOS, runtime.GOARCH)
		url  = fmt.Sprintf("%s/%s", goBinaryHost, tar)
		tmp  = filepath.Join(v.wsconf.WorkingDir, ".go", "release")
		file = filepath.Join(tmp, tar)
	)
	if fs.Exists(bin) {
		return nil
	}
	fs.Mkdir(path)
	fs.Mkdir(tmp)

	// downloading go release tar.gz
	v.status.Start(fmt.Sprintf("[Go] Downloading prebuilt release %s (%s/%s)...", v.goconf.Version, runtime.GOOS, runtime.GOARCH))
	out, err := os.Create(file)
	if err != nil {
		v.status.End(false)
		return err
	}
	defer out.Close()
	rsp, err := http.Get(url)
	if err != nil {
		v.status.End(false)
		return err
	}
	defer rsp.Body.Close()
	_, err = io.Copy(out, rsp.Body)
	if err != nil {
		v.status.End(false)
		return err
	}
	v.status.End(true)

	// unpacking files to workspace
	v.status.Start(fmt.Sprintf("[Go] Unpacking binaries..."))
	if err := archiver.NewTarGz().Unarchive(file, path); err != nil {
		v.status.End(false)
		return err
	}
	v.status.End(true)

	return nil
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

func (v *plugin) Bin() []string {
	return []string{filepath.Join(v.wsconf.WorkingDir, ".go", v.goconf.Version, "go", "bin")}
}

// Export is a plugin instance used for workspace
var Export = ext.Extension(new(plugin))
