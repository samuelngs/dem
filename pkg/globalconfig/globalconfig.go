package globalconfig

import (
	"io/ioutil"
	"os"

	"github.com/samuelngs/workspace/pkg/util/homedir"
	"gopkg.in/yaml.v2"
)

// Settings is a GlobalConfig instance used for convienience
var Settings = &GlobalConfig{
	StorageDir: homedir.Path("workspace"),
	PluginsDir: homedir.Path(".config/workspace/plugins"),
}

// GlobalConfig is the global configuration
type GlobalConfig struct {
	// the root storage path of virtual workspaces
	StorageDir string `yaml:"storage_dir"`

	// the plugin path, command line tool would load the
	// `so` modules defined in workspace configuration
	PluginsDir string `yaml:"plugins_dir"`
}

// Load reads and parses global workspace configuration from yaml file
func Load(cfgPath string) error {
	dat, err := ioutil.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if err := yaml.Unmarshal(dat, Settings); err != nil {
		return err
	}
	return nil
}
