package main

import (
	"bufio"
	"os"
	"regexp"
)

// A single line within a logfile
type LogLine struct {
	AppName      string
	LogPath      string
	Description  string
	ErrorLevel   string
	TimeStamp    string
	SourceIP     string
	Request      string
	Code         string
	UserAgent    string
	ResponseSize int64
	Referrer     string
	Upstream     string
}

// Represents a log file, e.g nginx.log
type LogFile struct {
	AppName           string   `yaml:"name"`
	Filepath          string   `yaml:"filepath"`
	AlertInterval     string   `yaml:"alert-interval"`
	CaptureConditions []string `yaml:"capture-line-if"`
	LogType           string   `yaml:"type"`
	LastTimestamp     string
	LastByteRead      int64
	Regex             string
}

// TODO:
// Keep the log file looping
// Open the log file from where I last stopped
// Convert the timestamp format
// Add support for access log
// Notes:
// https://github.com/Knetic/govaluate
// https://github.com/oleksandr/conditions

// Used to stop the monitoring threads neatly
var stopLogMonitoring bool

//
func StartLogMonitoring(settings *Settings, loglines chan LogLine) {

	err, regexes := LoadLogFileRegex()
	if err != nil {
		panic("Cannot load log file regex")
	}

	for _, logFile := range settings.LogFiles {

		// We get the parsing regex for this filetype from
		// the log_formats.yaml file. Using this method, it's easy to
		// add a new format, just define it in log_formats and then
		// specify the name of the newly defined one in 'type'
		logFile.Regex = regexes[logFile.LogType]

		// Make sure logtype is set. If it's not, no point parsing as we can't
		// get the values anyways.
		if len(logFile.Regex) <= 0 {
			print("No LogType was specified for this log. Cannot monitor")
			continue
		}

		go monitorLog(logFile, loglines)
	}
}

func monitorLog(logFile LogFile, loglines chan LogLine) {

	// Open the log file
	f, err := os.Open(logFile.Filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close() // close at end of function

	// Open with scanner so we can check each line
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	// Retrieve the regular expression we will use to parse the line
	var expression = regexp.MustCompile(logFile.Regex)

	// Loop through each line in the file
	for scanner.Scan() {

		// Structure where we will save the line
		var logline LogLine

		// Get some info from the file itself
		logline.LogPath = logFile.Filepath
		logline.AppName = logFile.AppName

		// Find the matching text in the log
		match := expression.FindStringSubmatch(scanner.Text())

		// Get each value
		for i, name := range expression.SubexpNames() {
			if name == "errorlevel" {
				logline.ErrorLevel = match[i]
			} else if name == "description" {
				logline.Description = match[i]
			} else if name == "dateandtime" {
				logline.TimeStamp = match[i]
			}
		}

		loglines <- logline
	}

}

func StopReadingLogs() {
	stopLogMonitoring = true
}
