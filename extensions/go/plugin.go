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
	"github.com/samuelngs/workspace/pkg/util/envcomposer"
	"github.com/samuelngs/workspace/pkg/util/fs"
	"github.com/samuelngs/workspace/pkg/workspaceconfig"
	"gopkg.in/yaml.v2"
)

var goBinaryHost = "https://dl.google.com/go"

type plugin struct {
	wsconf *workspaceconfig.Config
	goconf *goConfig
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
	return true, nil
}

func (v *plugin) StartPre() error {
	var (
		path = filepath.Join(v.wsconf.WorkingDir, ".go", v.goconf.Version)
		tar  = fmt.Sprintf("go%s.%s-%s.tar.gz", v.goconf.Version, runtime.GOOS, runtime.GOARCH)
		url  = fmt.Sprintf("%s/%s", goBinaryHost, tar)
		tmp  = filepath.Join(os.TempDir(), tar)
	)
	if fs.Exists(path) {
		return nil
	}
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer out.Close()
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	_, err = io.Copy(out, rsp.Body)
	if err != nil {
		return err
	}
	return archiver.NewTarGz().Unarchive(tmp, path)
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
