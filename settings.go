package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config struct for webapp config
type Settings struct {
	ContainerName        string `yaml:"container-name"`
	ContainerSupport     string `yaml:"container-support"`
	ContainerDescription string `yaml:"container-description"`

	LogFiles []LogFile `yaml:"logs"`

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

	// Set sensible defaults for uptime list
	for i := 0; i < len(settings.UptimeRequestList); i++ {

		if len(settings.UptimeRequestList[i].CheckInterval) <= 0 {
			settings.UptimeRequestList[i].CheckInterval = "5m" // 5 minutes
		}

		fmt.Printf("Request to monitor endpoint: " + settings.UptimeRequestList[i].Endpoint + " @ " + settings.UptimeRequestList[i].CheckInterval + "\n")
	}

	// Set sensible defaults for logs
	for i := 0; i < len(settings.LogFiles); i++ {

		// If no alert interval is set, set it to 15 minutes. This is
		// the default alert interval, which can be changed on a per-item
		// basis in the CaptureConditions.
		if len(settings.LogFiles[i].AlertInterval) <= 0 {
			settings.LogFiles[i].AlertInterval = "15m" // 15 minutes
		}

		_, err := os.Stat(settings.LogFiles[i].Filepath)
		if os.IsNotExist(err) {
			print("WARNING: File " + settings.LogFiles[i].Filepath + " does not exist!")
		}

		fmt.Printf("Request to monitor logfile: " + settings.LogFiles[i].Filepath + " @ " + settings.LogFiles[i].AlertInterval + "\n")
	}

	return settings, nil
}

func LoadLogFileRegex() (*LogFileRegex, error) {

	logRegex := &LogFileRegex{}

	// Open config file
	file, err := os.Open("./log_formats.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&logRegex); err != nil {
		return nil, err
	}

	return logRegex, nil
}

func print(str string) {
	fmt.Println(str)
}
