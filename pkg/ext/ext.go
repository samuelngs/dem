package ext

import "github.com/samuelngs/workspace/pkg/workspaceconfig"

// Extension interface
type Extension interface {
	Init(*workspaceconfig.Config) (bool, error)
	StartPre() error
	Environment() map[string]string
	Aliases() map[string]string
	Bin() []string
}
