package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config struct for webapp config
type Settings struct {
	ContainerName        string `yaml:"container_name"`
	ContainerSupport     string `yaml:"container_support"`
	ContainerDescription string `yaml:"container_description"`

	Nginx struct {
		Logfilename string `yaml:"log_filename"`
	} `yaml:"nginx"`

	Uptime struct {
		Endpoint           string `yaml:"endpoint"`
		GeneralStatusCheck string `yaml:"general_status_check"`
		CheckInterval      string `yaml:"check_interval"`
	} `yaml:"uptime"`
}

func LoadSettings(settingsFile string) (*Settings, error) {

	settings := &Settings{}

	// Open config file
	file, err := os.Open(settingsFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&settings); err != nil {
		return nil, err
	}

	// Set sensible defaults
	if len(settings.Uptime.CheckInterval) <= 0 {
		settings.Uptime.CheckInterval = "5m" // 5 minutes
	}

	return settings, nil
}
