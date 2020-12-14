package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
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
	StatusCode   string
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
	Regex             string   // Loaded from log_formats.yaml file
	LastTimestamp     string   // This is persisted in the lorona.dat file
	LastByteRead      int64    // This is persisted in the lorona.dat file
	LogFirstFewLines  string   // This is persisted in the lorona.dat file
}

// TODO:
// Keep the log file looping
// Open the log file from where I last stopped
// Convert the timestamp format
// Notes:
// https://github.com/Knetic/govaluate
// https://github.com/oleksandr/conditions

// Used to stop the monitoring threads neatly
var stopLogMonitoring bool

// Start the threads that will monitor each log
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

		// Start the go-routine that will be monitoring the logs
		go monitorLog(logFile, loglines)
	}
}

func StopReadingLogs() {
	stopLogMonitoring = true
}

// Problems
// - Very large files - solved
// - Files rotated - solved
// - Files deleted
// -
func monitorLog(logFile LogFile, loglines chan LogLine) {

	// Open the log file
	f, err := os.Open(logFile.Filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close() // close at end of function

	fi, err3 := f.Stat()
	if err3 != nil {
		panic(err3)
	}

	if fi.Size() > 100 {

		// Create a buffer we will store the first few lines of the log
		FirstFewLines := make([]byte, 100)

		_, err := f.Read(FirstFewLines)
		if err != nil {
			fmt.Println(err)
		}

		FirstFewLinesS := string(FirstFewLines)

		if len(logFile.LogFirstFewLines) > 0 && logFile.LogFirstFewLines != FirstFewLinesS {
			// this case means the start of the log file has changed, so the log has likely
			// been rotated. We discard our existing info about our position and start from
			// the beginning again.
			logFile.LastByteRead = 0
		}

		logFile.LogFirstFewLines = string(FirstFewLines)

		// Get the unread portion of the log
		var lengthToRead = fi.Size() - logFile.LastByteRead
		fmt.Printf("Unread portion of log is %d", lengthToRead)

		// If the pending length of the log exceeds 1MB, we will skip over
		// a big part of the log and start reading at the last 1MB
		if lengthToRead > 1000000 {
			logFile.LastByteRead = fi.Size() - 1000000
		}

		// Seek to the position last read
		_, err2 := f.Seek(logFile.LastByteRead, 0)
		if err2 != nil {
			panic(err2)
		}
	}

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
		if len(match) == 0 {
			continue
		}

		// Get each value
		for i, name := range expression.SubexpNames() {
			if name == "errorlevel" {
				logline.ErrorLevel = match[i]
			} else if name == "description" {
				logline.Description = match[i]
			} else if name == "timestamp" {
				logline.TimeStamp = match[i]

				// We will need to check if this log has already been read (earlier than 'lastread')
				// If so, we continue
			} else if name == "ipaddress" {
				logline.SourceIP = match[i]
			} else if name == "statuscode" {
				logline.StatusCode = match[i]
			} else if name == "useragent" {
				logline.UserAgent = match[i]
			} else if name == "referrer" {
				logline.Referrer = match[i]
			} else if name == "bytessent" {
				logline.ResponseSize, _ = strconv.ParseInt(match[i], 10, 64) // convert to a 64bit int in base 10
			}
		}

		loglines <- logline
	}

}
