package ext

import (
	"github.com/samuelngs/dem/pkg/workspaceconfig"
)

// Extension interface
type Extension interface {
	Init(*workspaceconfig.Config) (bool, error)
	SetupTasks() SetupTasks
	Environment() map[string]string
	Aliases() map[string]string
	Paths() []string
	String() string
}
