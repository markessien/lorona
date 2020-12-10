package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Config struct for webapp config
type Settings struct {
	Nginx struct {
		Logfilename string `yaml:"log_filename"`
	} `yaml:"nginx"`
}

func loadSettings(settingsFile string) (*Settings, error) {

	settings := &Settings{}

	// Open config file
	file, err := os.Open(configPath)
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

	return settings, nil
}

func main() {
	fmt.Println("Hello, World!")

	settings, err := loadSettings("settings.yaml")
	if err != nil {
		log.Fatal(err)
	}
}
