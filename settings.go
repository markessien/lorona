package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type UptimeRequest struct {
	Endpoint       string `yaml:"url"`
	StatusCheck    string `yaml:"status_check"`
	ExpectedStatus string `yaml:"expected_status"`
	CheckInterval  string `yaml:"check_interval"`
	GetTokenUrl    string `yaml:"get_token_url"`
}

// Config struct for webapp config
type Settings struct {
	ContainerName        string `yaml:"container_name"`
	ContainerSupport     string `yaml:"container_support"`
	ContainerDescription string `yaml:"container_description"`

	Nginx struct {
		ErrorLogfilename string `yaml:"error_log_filename"`
	} `yaml:"nginx"`

	UptimeRequestList []UptimeRequest `yaml:"uptime"`
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
	for i := 0; i < len(settings.UptimeRequestList); i++ {

		if len(settings.UptimeRequestList[i].CheckInterval) <= 0 {
			settings.UptimeRequestList[i].CheckInterval = "5m" // 5 minutes
		}

		fmt.Printf("Request to monitor endpoint: " + settings.UptimeRequestList[i].Endpoint + " @ " + settings.UptimeRequestList[i].CheckInterval + "\n")
	}

	return settings, nil
}

func print(str string) {
	fmt.Println(str)
}
