package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Repository struct {
	URL    string `yaml:"url"`
	Branch string `yaml:"branch"`
}

type RunAs struct {
	User  string `yaml:"user"`
	Group string `yaml:"group"`
}

type ServiceConfig struct {
	Name           string     `yaml:"name"`
	Repository     Repository `yaml:"repository"`
	DeploymentsDir string     `yaml:"deployments_dir"`
	Executable     string     `yaml:"executable"`
	PreStartHook   string     `yaml:"pre_start_hook"`
	ListenPort     int        `yaml:"listen_port"`
	TargetPorts    []int      `yaml:"target_ports"`
	RunAs          RunAs      `yaml:"run_as"`
}

type Config struct {
	WebhookPort int                        `yaml:"webhook_port"`
	Services    map[string][]ServiceConfig `yaml:"services"`
}

// Loads the config and returns the Config struct
func Load() (*Config, error) {
	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Helper to get a service by name
func (c *Config) GetService(name string) (*ServiceConfig, error) {
	services, ok := c.Services[name]
	if !ok || len(services) == 0 {
		return nil, os.ErrNotExist
	}
	return &services[0], nil
}
