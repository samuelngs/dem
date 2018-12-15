package bash

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/samuelngs/dem/pkg/util/exec"
	"github.com/samuelngs/dem/pkg/util/fs"
)

var bashrc, _ = template.New("bashrc").Parse(`
{{- $homedir := .Home -}}
{{- $extension_bin := .ExtensionBin -}}
{{- $aliases := .Aliases -}}
{{- $sources := .Sources -}}
{{- $envvars := .EnvironmentVariables -}}

if [ -f "{{$homedir}}/.bashrc" ]; then
  source {{$homedir}}/.bashrc
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

// Bash accepts a --rcfile filename option (custom ~/.bashrc). Run function would
// inject and pass custom run command script to initial interactive shell.
func (v *bash) Run() error {
	var (
		b               bytes.Buffer
		homedir         = v.GetEnv("HOME")
		dotdir          = filepath.Join(homedir, ".workspace_shell")
		bashrcPath      = filepath.Join(dotdir, ".bashrc")
		sudoWarningPath = filepath.Join(homedir, ".sudo_as_admin_successful")
	)
	opts := &options{
		Home:                 homedir,
		ExtensionBin:         v.GetEnv("EXT_PATH"),
		Aliases:              v.GetAliases(),
		Sources:              v.GetSources(),
		EnvironmentVariables: v.GetEnvs(),
	}

	// write bash startup files
	fs.Mkdir(dotdir)
	if err := bashrc.Execute(&b, opts); err != nil {
		return err
	}
	fs.WriteFile(bashrcPath, b.Bytes())

	// hide bash sudo as root warning
	// -------------------------------------------------------------------------
	// | To run a command as administrator (user "root"), use "sudo <command>".
	// | See "man sudo_root" for details.
	// -------------------------------------------------------------------------
	fs.WriteFile(sudoWarningPath, []byte(""))

	// override arguments
	v.Command.SetArgs("--rcfile", bashrcPath)

	return v.Command.Run()
}

// New initializes bash version of exec command
func New(command string, args ...string) exec.Command {
	return &bash{exec.New(command, args...)}
}
