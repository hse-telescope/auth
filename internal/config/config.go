package config

import (
	"log"
	"os"

	"github.com/hse-telescope/utils/db/psql"
	"github.com/hse-telescope/utils/queues/kafka"
	"gopkg.in/yaml.v3"
)

type Clients struct{}

// Config ...
type Config struct {
	Port      uint16                 `yaml:"port"`
	DB        psql.DB                `yaml:"db"`
	Clients   Clients                `yaml:"clients"`
	JWTSecret string                 `yaml:"jwt_secret"`
	Kafka     kafka.QueueCredentials `yaml:"queue_credentials"`
}

// Parse ...
func Parse(path string) (Config, error) {
	bytes, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		return Config{}, err
	}

	log.Default().Println("\n---START PARSING---")
	config := Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Default().Printf("\n---CONFIG ERR---\n[ERR]: %s\n", err.Error())
		return Config{}, err
	}
	log.Default().Println("\n---FINISH PARSING---")

	log.Default().Println("\n---PARSED CONFIG---\n[CONFIG]:\n", config)

	return config, nil
}
