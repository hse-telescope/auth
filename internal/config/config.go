package config

import (
	"os"

	"github.com/hse-telescope/utils/db/psql"
	"gopkg.in/yaml.v3"
)

type Clients struct{}

// Config ...
type Config struct {
	Port    uint16  `yaml:"port"`
	DB      psql.DB `yaml:"db"`
	Clients Clients `yaml:"clients"`
}

// Parse ...
func Parse(path string) (Config, error) {
	bytes, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
