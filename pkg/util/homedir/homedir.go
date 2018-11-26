package homedir

import (
	"fmt"

	"github.com/samuelngs/dem/pkg/util/env"
)

// Dir returns home directory path
func Dir() string {
	return env.GetEnvAsStringWithFallback("UNMASK_HOME", "HOME")
}

// Path composer
func Path(s string) string {
	return fmt.Sprintf("%s/%s", Dir(), s)
}
