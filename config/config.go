package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repository struct {
		URL    string `yaml:"url"`
		Branch string `yaml:"branch"`
	} `yaml:"repository"`

	Service struct {
		Name           string `yaml:"name"`
		DeploymentsDir string `yaml:"deployments_dir"`
		Executable     string `yaml:"executable"`
		PreStartHook   string `yaml:"pre_start_hook"`
		ListenPort     int    `yaml:"listen_port"`
		TargetPorts    []int  `yaml:"target_ports"`
	} `yaml:"service"`
}

func Load() (*Config, error) {
	file, err := os.Open("/etc/autopuller/config.yaml")
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
