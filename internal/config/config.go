package config

import (
	"os"

	"github.com/hse-telescope/utils/db/psql"
	"gopkg.in/yaml.v3"
)

type Clients struct{}

// Kafka config...
type KafkaConfig struct {
	URLs  []string `yaml:"urls"`
	Topic string   `yaml:"topic"`
}

// Config ...
type Config struct {
	Port      uint16      `yaml:"port"`
	DB        psql.DB     `yaml:"db"`
	Clients   Clients     `yaml:"clients"`
	JWTSecret string      `yaml:"jwt_secret"`
	Kafka     KafkaConfig `yaml:"queue_credentials"`
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
