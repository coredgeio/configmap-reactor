package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type configType struct {
	Label string `yaml:"label"`
}

var config configType

func GetLabel() string {
	return config.Label
}

// ParseConfig returns a new decoded Config struct
func ParseConfig(configPath string) error {
	// validate config path before decoding
	if err := validateConfigPath(configPath); err != nil {
		return err
	}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return err
	}

	return nil
}
