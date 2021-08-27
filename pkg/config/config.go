package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Phys     string                   `yaml:"phys"`
	Triggers map[uint16]CommandConfig `yaml:"triggers"`
}

type CommandConfig struct {
	Command  Command
	Interval time.Duration `yaml:"interval"`
}

type Command []string

func Read(name string) (*Config, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("reading file failed: %w", err)
	}
	var c *Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("unmarshal yaml failed: %w", err)
	}
	return c, nil
}
