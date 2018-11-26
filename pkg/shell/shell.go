package shell

import (
	"path/filepath"

	"github.com/samuelngs/dem/pkg/shell/zsh"
	"github.com/samuelngs/dem/pkg/util/exec"
)

// New initializes exec command
func New(command string, args ...string) exec.Command {
	switch filepath.Base(command) {
	case "zsh":
		return zsh.New(command, args...)
	default:
		return exec.New(command, args...)
	}
}
