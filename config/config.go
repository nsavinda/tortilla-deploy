package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repo struct {
		CloneURL string `yaml:"clone_url"`
		Branch   string `yaml:"branch"`
	} `yaml:"repo"`

	Service struct {
		Name         string `yaml:"name"`
		ClonePath    string `yaml:"clonePath"`
		ExecCommand  string `yaml:"execCommand"`
		PreStartHook string `yaml:"preStartHook"`
	} `yaml:"service"`
}

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
