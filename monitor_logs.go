package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/araddon/dateparse"
)

type LogSummary struct {
	StatusCount        map[string]int64
	SeverityLevelCount map[string]int64
}

// A single line within a logfile
type LogLine struct {
	// Basics
	AppName string
	LogPath string

	// Natively supported log entry types
	// These are also auto exposed
	Description     string
	Severity        string
	TimeStampString string
	TimeStamp       time.Time
	StatusCode      string
	ExecutionTime   uint64

	// All other log items have to be parsed
	// via regex.
	/*
		SourceIP        string
		Request         string
		UserAgent       string
		ResponseSize    int64
		Referrer        string
		Upstream        string
	*/

	Fields map[string]interface{}
}

// Represents a log file, e.g nginx.log
type LogFile struct {
	AppName           string   `yaml:"name"`
	Filepath          string   `yaml:"filepath"`
	AlertInterval     string   `yaml:"alert-interval"`
	CaptureConditions []string `yaml:"capture-line-if"`
	LogType           string   `yaml:"type"`
	TimeFormatName    string   `yaml:"time-format"`
	TimeFormat        string   // Loaded from log_formats.yaml file
	Regex             string   // Loaded from log_formats.yaml file
	LastTimestamp     string   // This is persisted in the lorona.dat file
	LastByteRead      int64    // This is persisted in the lorona.dat file
	LogFirstFewLines  string   // This is persisted in the lorona.dat file
}

// TODO:
// Keep the log file looping
// Run the line evaluation to decide if to keep it
// Check the timestamp to make sure it is newer than last we had
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
		logFile.TimeFormat = regexes[logFile.TimeFormatName]

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

// Monitors log files.
// Part 1: Make sure we do not load massive log files into memory
// Part 2: Evaluate log file line capture conditions
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

	// Confirm that it's the same file by looking at first few lines
	// If the filesize is larger than 100, we check for the signature
	// at the start. This is used to know if the log was rotated or
	// changed
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
		logline.Fields = make(map[string]interface{})

		// Get some info from the file itself
		logline.LogPath = logFile.Filepath
		logline.AppName = logFile.AppName

		// Find the matching text in the log
		match := expression.FindStringSubmatch(scanner.Text())
		if len(match) == 0 {
			continue
		}

		// If set to true, this log line will not be added to the final log
		var ignore_this_line = false

		// This is where all the values for all the fields will be stored. This can be used
		// for the evaluation of the condition if this particular line should be added to
		// the log
		condition_parameters := make(map[string]interface{}, 8)

		// Get each value
		for i, name := range expression.SubexpNames() {
			if name == "severity" {
				logline.Severity = match[i]
			} else if name == "description" {
				logline.Description = match[i]
			} else if name == "timestamp" {

				logline.TimeStampString = match[i]

				if len(logFile.TimeFormat) > 0 {
					t, err := time.Parse(logFile.TimeFormat, logline.TimeStampString)
					if err == nil {
						logline.TimeStamp = t
					}
				} else {
					// We try a freestyle timestamp parsing
					t, err := dateparse.ParseAny(logline.TimeStampString)
					if err == nil {
						logline.TimeStamp = t
					}
				}

				condition_parameters["time_timestamp"] = logline.TimeStamp

			} else if name == "statuscode" {
				logline.StatusCode = match[i]
				condition_parameters["int_statuscode"], _ = strconv.Atoi(match[i])
			} else if name == "executiontime" {
				condition_parameters["int_executiontime"], _ = strconv.Atoi(match[i])
			} else if len(name) > 0 {
				// One of the non-default keys came. We put it in the map
				logline.Fields[name] = match[i]

				// We also put int, float and time versions
				if intVal, err := strconv.ParseInt(match[i], 10, 64); err == nil {
					logline.Fields["int_"+name] = intVal
				}

			}

			condition_parameters[name] = match[i]
		}

		// We run the evaluator to figure out if we need to even add this line to the logs
		// The user can specify conditions in the settings yaml file for when a log should
		// be captured. We use a generic evaluator, which creates maximum flexibility for
		// the user

		for _, condition := range logFile.CaptureConditions {

			ignore_this_line = true
			// condition is something like "statuscode = 404" or "statuscode > 500 AND statuscode < 599 THEN alert immediately"

			// First we remove the 'then' part as it's not part of the conditional
			then_pos := strings.Index(strings.ToLower(condition), "then")
			if then_pos > 0 {
				then_part := condition[then_pos:len(condition)]
				condition = condition[0 : then_pos-1] // remove the THEN part from condition
				print("Then condition not working yet " + then_part)
			}

			expression, err := govaluate.NewEvaluableExpression(condition)
			if err != nil {
				print("Could not evaluate expression " + condition)
				continue
			}

			// We have the parameters and their value, so we can now run the specified conditional
			// to know if this line should be added or not
			result, err := expression.Evaluate(condition_parameters)
			// result is now set to "true", the bool value.
			if err == nil && result == true {
				ignore_this_line = false
				break
			}
		}

		if ignore_this_line == false {
			loglines <- logline
		}

	}

}
