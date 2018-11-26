package exec

import (
	"io"
	"os"
	"os/exec"
)

// Command abstracts over creating command
type Command interface {
	Run() error
	SetDir(string)
	SetEnv(...string)
	SetStdin(io.Reader)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
}

type command struct {
	dir            string
	cmd            string
	args           []string
	envs           []string
	stdin          io.Reader
	stdout, stderr io.Writer
	startPre       []func() error
	startPost      []func() error
	stopPost       []func() error
}

func (v *command) Run() error {
	cmd := exec.Command(v.cmd, v.args...)
	cmd.Dir = v.dir
	cmd.Env = v.envs
	cmd.Stdin = v.stdin
	cmd.Stdout = v.stdout
	cmd.Stderr = v.stderr
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

func (v *command) SetEnv(envs ...string) {
	v.envs = append(v.envs, envs...)
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

// New creates abstracted command interface
func New(cmd string, args ...string) Command {
	c := &command{
		cmd:    cmd,
		args:   args,
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	return c
}
