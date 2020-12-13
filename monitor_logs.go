package main

import (
	"bufio"
	"os"
	"regexp"
)

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

type LogFile struct {
	AppName           string   `yaml:"name"`
	Filepath          string   `yaml:"filepath"`
	AlertInterval     string   `yaml:"alert-interval"`
	CaptureConditions []string `yaml:"capture-line-if"`
	LogType           string
	LastTimestamp     string
	LastByteRead      int64
	Regex             string
}

type LogFileRegex struct {
	NginxErrorRegex  string `yaml:"nginx_error_log"`
	NginxAccessRegex string `yaml:"nginx_access_log"`
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
var logFileRegex LogFileRegex

//
func StartLogMonitoring(settings *Settings, loglines chan LogLine) {

	/*
		logFileRegex, err := LoadLogFileRegex()
		if err != nil {
			panic("Cannot load log file regex")
		}

		// Each log file is set differently and has a different format.
		// We look at which is set and start it up separately
		if len(settings.Nginx.ErrorLogfilename) > 0 {

			var logFile LogFile
			logFile.AppName = "NGINX"
			logFile.LogType = "ErrorLog"
			logFile.Filepath = settings.Nginx.ErrorLogfilename
			logFile.Regex = logFileRegex.NginxErrorRegex

			print("Monitoring Nginx Error logfile at " + logFile.Filepath)

			if _, err := os.Stat(logFile.Filepath); err == nil {
				go monitorLog(logFile, loglines)

			} else if os.IsNotExist(err) {
				print("Nginx logfile does not exist")
			}
		}
	*/
}

func monitorLog(logFile LogFile, loglines chan LogLine) {
	f, err := os.Open(logFile.Filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {

		var logline LogLine
		logline.LogPath = logFile.Filepath
		logline.AppName = logFile.AppName

		var expression = regexp.MustCompile(logFile.Regex)
		match := expression.FindStringSubmatch(scanner.Text())

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
