package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Command abstracts over creating command
type Command interface {
	Run() error
	SetDir(string)
	SetEnv(map[string]string)
	SetAliases(map[string]string)
	SetStdin(io.Reader)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	GetEnv(string) string
	GetAliases() map[string]string
}

type command struct {
	dir            string
	cmd            string
	args           []string
	envs           map[string]string
	aliases        map[string]string
	stdin          io.Reader
	stdout, stderr io.Writer
}

func (v *command) Run() error {
	var i int
	cmd := exec.Command(v.cmd, v.args...)
	cmd.Dir = v.dir
	cmd.Stdin = v.stdin
	cmd.Stdout = v.stdout
	cmd.Stderr = v.stderr
	cmd.Env = make([]string, len(v.envs))
	for key, val := range v.envs {
		cmd.Env[i] = fmt.Sprintf("%s=%s", key, val)
		i++
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

func (v *command) SetCommand(cmd string) {
	v.cmd = cmd
}

func (v *command) SetArgs(args ...string) {
	v.args = args
}

func (v *command) SetDir(dir string) {
	v.dir = dir
}

func (v *command) SetEnv(env map[string]string) {
	for key, val := range env {
		v.envs[key] = val
	}
}

func (v *command) SetAliases(aliases map[string]string) {
	for alias, cmd := range aliases {
		v.aliases[alias] = cmd
	}
}

func (v *command) SetStdin(reader io.Reader) {
	v.stdin = reader
}

func (v *command) SetStdout(writer io.Writer) {
	v.stdout = writer
}

func (v *command) SetStderr(writer io.Writer) {
	v.stderr = writer
}

func (v *command) GetEnv(key string) string {
	return v.envs[key]
}

func (v *command) GetAliases() map[string]string {
	return v.aliases
}

// New creates abstracted command interface
func New(cmd string, args ...string) Command {
	c := &command{
		cmd:     cmd,
		args:    args,
		envs:    make(map[string]string),
		aliases: make(map[string]string),
		stdin:   os.Stdin,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
	}
	return c
}
