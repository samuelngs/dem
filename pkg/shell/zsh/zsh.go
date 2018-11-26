package zsh

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/samuelngs/workspace/pkg/util/exec"
	"github.com/samuelngs/workspace/pkg/util/fs"
)

var zshrc, _ = template.New("zshrc").Parse(`
{{- $homedir := .Home -}}
{{- $extension_bin := .ExtensionBin -}}
{{- $aliases := .Aliases -}}

if [ -f "{{$homedir}}/.zshrc" ]; then
  source {{$homedir}}/.zshrc
fi

if [ ! -z "{{$extension_bin}}" ]; then
  export PATH={{$extension_bin}}:$PATH
fi

{{- range $alias, $command := $aliases}}
alias {{$alias}}="{{$command}}"
{{end}}
`)

type options struct {
	Home         string
	ExtensionBin string
	Aliases      map[string]string
}

type zsh struct {
	exec.Command
}

func (v *zsh) Run() error {
	var (
		b         bytes.Buffer
		homedir   = v.GetEnv("HOME")
		dotdir    = filepath.Join(homedir, ".workspace_zsh")
		zshrcPath = filepath.Join(dotdir, ".zshrc")
		symlinks  = []string{".zprofile", ".zshenv", ".zlogin", ".zlogout"}
	)

	opts := &options{
		Home:         homedir,
		ExtensionBin: v.GetEnv("EXT_PATH"),
		Aliases:      v.GetAliases(),
	}

	// write zsh startup files
	fs.Mkdir(dotdir)
	for _, file := range symlinks {
		path := filepath.Join(homedir, file)
		dest := filepath.Join(dotdir, file)
		fs.Symlink(path, dest)
	}

	if err := zshrc.Execute(&b, opts); err != nil {
		return err
	}
	fs.WriteFile(zshrcPath, b.Bytes())

	// set zsh dotfile path
	v.SetEnv(map[string]string{
		"ZDOTDIR": dotdir,
	})

	return v.Command.Run()
}

// New initializes zsh version of exec command
func New(command string, args ...string) exec.Command {
	return &zsh{exec.New(command, args...)}
}