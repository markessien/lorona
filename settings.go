package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// This contains the result format that will be converted to json
// and sent to the remove receiver.
type Results struct {
	ContainerName        string
	ContainerSupport     string
	ContainerDescription string
	FileFormat           string
	UptimeList           []UptimeResponse
	LoglineList          []LogLine
	SysMonitorInfoList   []SysMonitorInfo
}

// Config structure for the requests to this app that the user has
type Settings struct {
	ContainerName        string                 `yaml:"container-name"`
	ContainerSupport     string                 `yaml:"container-support"`
	ContainerDescription string                 `yaml:"container-description"` // A user set description of what this container (or system) is all about
	DataFile             string                 `yaml:"data-file"`             // The location of the data file where we will store resumption points for logs
	LogFiles             []LogFile              `yaml:"logs"`                  // Requests for the log files we want to monitor
	UptimeRequestList    []UptimeRequest        `yaml:"uptime"`                // Contains all endpoints to be monitored
	SysMonitorRequest    SystemMonitorRequest   `yaml:"system"`                // Requests for the system parameters we want to monitor
	BackupMonitorRequest []BackupMonitorRequest `yaml:"backups-monitor"`       // Requests for the log files we want to monitor
	ObservedBackupFiles  []string               // This is where we store the backup files we have seen in our backup folders already
}

// Load settings from the settings.yaml file. All the settings are taken
// from the YAML files and put into the Settings structure above. Also, if
// some required things are not set, we assign sensible defaults here too.
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

	LoadData(settings)

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

	SaveData(settings)

	return settings, nil
}

// This function loads the log_formats.yaml file and stores all
// the regex parsers for each log file format in a key value.
// Any new type added to the file will be inserted in there.
func LoadLogFileRegex() (error, map[string]string) {

	// Open file with all the regexes
	file, err := os.Open("./log_formats.yaml")
	if err != nil {
		return err, nil
	}
	defer file.Close()

	// read the file using the scanner
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	// The regexes will be stored here as key:value
	var regexes = make(map[string]string)

	// loop through each line
	for scanner.Scan() {

		// Get this single regex line
		var regex_line = scanner.Text()

		// Make sure it's not a comment and it is valid
		if len(regex_line) > 3 && !strings.HasPrefix(regex_line, "#") && strings.Index(regex_line, ":") > 1 {

			// Split, using the : as the delimeter. SplitN forces split to just 2 groups
			var items = strings.SplitN(regex_line, ":", 2)
			if len(items) == 2 {

				// Save in the key-value, remove leading spaces and quotes
				regexes[strings.TrimSpace(items[0])] = strings.Trim(strings.TrimSpace(items[1]), "\"'")
			}

		}
	}

	return nil, regexes
}

// Print ersatz
func print(str string) {
	fmt.Println(str)
}

// Loads the last settings file. We need it for some stuff
// like info about the log files
func LoadData(settings *Settings) error {

	dataFile, err := os.Open(settings.DataFile)
	if err != nil {
		fmt.Println(err)
		return err
	}

	dataSettings := Settings{}

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&dataSettings)

	if err != nil {
		fmt.Println(err)
		return err
	}

	// We transfer all the position info from the log files to the settings
	// structure. This position info is used to make sure we read from a pos
	// advanced in the file (efficiency)

	// Yes, I know we can do this better, but monitored log files should
	// be a small number, so a double loop should not matter.
	for _, logFileData := range dataSettings.LogFiles {

		for i := 0; i < len(settings.LogFiles); i++ {
			l := settings.LogFiles[i].Filepath
			if l == logFileData.Filepath {
				settings.LogFiles[i].LastTimestamp = logFileData.LastTimestamp
				settings.LogFiles[i].LastByteRead = logFileData.LastByteRead
				settings.LogFiles[i].LogFirstFewLines = logFileData.LogFirstFewLines
				break
			}
		}
	}

	dataFile.Close()
	return nil
}

// Saves our settings structure. We update our settings
// structure regularly with info like last read point in
// files, so this persists it, in case tool is restarted
func SaveData(settings *Settings) error {

	// create a file
	dataFile, err := os.Create(settings.DataFile)

	if err != nil {
		fmt.Println(err)
		return err
	}

	// serialize the data
	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(&settings)

	dataFile.Close()

	return nil
}
