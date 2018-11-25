package workspaceconfig

import (
	"io/ioutil"
	"os"

	"github.com/samuelngs/workspace/pkg/util/env"
	"gopkg.in/yaml.v2"
)

// WorkspaceConfig is the workspace configuration
type WorkspaceConfig struct {
	Env   map[string]string `yaml:"env"`
	Shell *Shell            `yaml:"shell"`
}

// Shell configuration
type Shell struct {
	Program string   `yaml:"program"`
	Args    []string `yaml:"args"`
}

// DefaultConfiguration returns default configuration
func DefaultConfiguration() *WorkspaceConfig {
	shell := &Shell{
		Program: env.GetEnvAsString("SHELL", "/bin/sh"),
		Args:    []string{"-l"},
	}
	conf := &WorkspaceConfig{
		Env:   make(map[string]string),
		Shell: shell,
	}
	return conf
}

// Load reads and parses workspace workspace configuration from yaml file
func Load(cfgPath string) (*WorkspaceConfig, error) {
	conf := DefaultConfiguration()
	dat, err := ioutil.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return conf, nil
	} else if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(dat, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// IsValid validates yaml configuration
func IsValid(cfgPath string) bool {
	_, err := Load(cfgPath)
	return err == nil
}

// New creates a new configuration with default settings
func New() ([]byte, error) {
	conf := DefaultConfiguration()
	return yaml.Marshal(conf)
}
