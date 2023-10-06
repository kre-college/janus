package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/hellofresh/janus/pkg/db"
	"github.com/hellofresh/janus/pkg/kafka"
)

type Config struct {
	DBUserManagement *db.Config    `yaml:"dbUserManagement"`
	DBRbac           *db.Config    `yaml:"dbRbac"`
	KafkaConfig      *kafka.Config `yaml:"kafkaConfig"`
}

func UnmarshalYAML(path string, dest *Config) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, dest)
}
