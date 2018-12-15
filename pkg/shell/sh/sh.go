package sh

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/samuelngs/dem/pkg/util/exec"
	"github.com/samuelngs/dem/pkg/util/fs"
)

var profile, _ = template.New("profile").Parse(`
{{- $homedir := .Home -}}
{{- $extension_bin := .ExtensionBin -}}
{{- $aliases := .Aliases -}}
{{- $sources := .Sources -}}
{{- $envvars := .EnvironmentVariables -}}

if [ "$SHELL" != "/bin/sh" ]; then
  exit 0
fi

if [ -f "{{$homedir}}/.profile_custom" ]; then
  source {{$homedir}}/.profile_custom
fi

if [ ! -z "{{$extension_bin}}" ]; then
  export PATH={{$extension_bin}}:$PATH
fi

{{- range $alias, $command := $aliases}}
alias {{$alias}}="{{$command}}"
{{end}}

{{- range $source := $sources}}
source {{$source}}
{{end}}

{{- range $key, $value := $envvars}}
export {{$key}}="{{$value}}"
{{end}}
`)

type options struct {
	Home                 string
	ExtensionBin         string
	Aliases              map[string]string
	Sources              []string
	EnvironmentVariables map[string]string
}

type bash struct {
	exec.Command
}

// inject and pass custom run command script to initial interactive shell.
func (v *bash) Run() error {
	var (
		b           bytes.Buffer
		homedir     = v.GetEnv("HOME")
		dotdir      = filepath.Join(homedir, ".workspace_shell")
		profilePath = filepath.Join(homedir, ".profile")
	)
	opts := &options{
		Home:                 homedir,
		ExtensionBin:         v.GetEnv("EXT_PATH"),
		Aliases:              v.GetAliases(),
		Sources:              v.GetSources(),
		EnvironmentVariables: v.GetEnvs(),
	}

	// write sh startup files
	fs.Mkdir(dotdir)
	if err := profile.Execute(&b, opts); err != nil {
		return err
	}
	fs.WriteFile(profilePath, b.Bytes())

	return v.Command.Run()
}

// New initializes sh version of exec command
func New(command string, args ...string) exec.Command {
	return &bash{exec.New(command, args...)}
}
