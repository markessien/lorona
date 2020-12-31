package main

import (
	"bufio"
	"encoding/gob"
	"io"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

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

// Configuration for logging
type LogConfig struct {
	ConsoleLoggingEnabled bool
	EncodeLogsAsJson      bool
	FileLoggingEnabled    bool
	Directory             string
	Filename              string
	MaxSize               int
	MaxBackups            int
	MaxAge                int
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

		lLog.Print("Request to monitor endpoint: " + settings.UptimeRequestList[i].Endpoint + " @ " + settings.UptimeRequestList[i].CheckInterval + "\n")
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
			lLog.Print("WARNING: File " + settings.LogFiles[i].Filepath + " does not exist!")
		}

		lLog.Print("Request to monitor logfile: " + settings.LogFiles[i].Filepath + " @ " + settings.LogFiles[i].AlertInterval + "\n")
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

// Loads the last settings file. We need it for some stuff
// like info about the log files
func LoadData(settings *Settings) error {

	dataFile, err := os.Open(settings.DataFile)
	if err != nil {
		lLog.Print(err)
		return err
	}

	dataSettings := Settings{}

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&dataSettings)

	if err != nil {
		lLog.Print(err)
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
		lLog.Print(err)
		return err
	}

	// serialize the data
	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(&settings)

	dataFile.Close()

	return nil
}

type Logger struct {
	*zerolog.Logger
}

// Configure sets up the logging framework
//
// In production, the container logs will be collected and file logging should be disabled. However,
// during development it's nicer to see logs as text and optionally write to a file when debugging
// problems in the containerized pipeline
//
// The output log file will be located at /var/log/service-xyz/service-xyz.log and
// will be rolled according to configuration set.
func ConfigureLogging(config LogConfig) *Logger {
	var writers []io.Writer

	if config.ConsoleLoggingEnabled {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}
	mw := io.MultiWriter(writers...)

	// zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(mw).With().Timestamp().Logger()

	logger.Info().
		Bool("fileLogging", config.FileLoggingEnabled).
		Bool("jsonLogOutput", config.EncodeLogsAsJson).
		Str("logDirectory", config.Directory).
		Str("fileName", config.Filename).
		Int("maxSizeMB", config.MaxSize).
		Int("maxBackups", config.MaxBackups).
		Int("maxAgeInDays", config.MaxAge).
		Msg("logging configured")

	return &Logger{
		Logger: &logger,
	}
}

func newRollingFile(config LogConfig) io.Writer {
	if err := os.MkdirAll(config.Directory, 0744); err != nil {
		log.Error().Err(err).Str("path", config.Directory).Msg("can't create log directory")
		return nil
	}

	return &lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.Filename),
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
	}
}
