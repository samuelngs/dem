package workspaceconfig

import (
	"io/ioutil"
	"os"

	"github.com/samuelngs/workspace/pkg/util/env"
	"gopkg.in/yaml.v2"
)

// Config is the root of configuration
type Config struct {
	Namespace  string     `yaml:"-"`
	WorkingDir string     `yaml:"-"`
	PluginsDir string     `yaml:"-"`
	Src        []byte     `yaml:"-"`
	Workspace  *Workspace `yaml:"workspace"`
}

// Workspace is the workspace configuration
type Workspace struct {
	Environment map[string]string      `yaml:"environment"`
	Aliases     map[string]string      `yaml:"aliases"`
	Shell       *Shell                 `yaml:"shell"`
	With        map[string]interface{} `yaml:"with"`
}

// Shell configuration
type Shell struct {
	Program string   `yaml:"program"`
	Args    []string `yaml:"args"`
}

// DefaultConfiguration returns default configuration
func DefaultConfiguration() *Config {
	shell := &Shell{
		Program: env.GetEnvAsString("SHELL", "/bin/sh"),
		Args:    []string{"-l"},
	}
	workspace := &Workspace{
		Environment: make(map[string]string),
		Aliases:     make(map[string]string),
		Shell:       shell,
	}
	conf := &Config{
		Workspace: workspace,
	}
	return conf
}

// Read returns workspace configuration yaml string
func Read(cfgPath string) ([]byte, error) {
	b, err := ioutil.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Parse parses workspace workspace configuration from yaml file
func Parse(dat []byte) (*Config, error) {
	conf := DefaultConfiguration()
	if err := yaml.Unmarshal(dat, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// IsValid validates yaml configuration
func IsValid(cfgPath string) bool {
	dat, err := Read(cfgPath)
	if err != nil {
		return false
	}
	_, err = Parse(dat)
	return err == nil
}

// New creates a new configuration with default settings
func New() ([]byte, error) {
	conf := DefaultConfiguration()
	return yaml.Marshal(conf)
}
