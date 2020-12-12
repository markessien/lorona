package main

import (
	"bufio"
	"fmt"
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
	filepath      string
	LastTimestamp string
	LastByteRead  int64
}

// Used to stop the monitoring threads neatly
var stopLogMonitoring bool

//
func StartLogMonitoring(settings *Settings, loglines chan LogLine) {

	// Each log file is set differently and has a different format.
	// We look at which is set and start it up separately
	if len(settings.Nginx.ErrorLogfilename) > 0 {
		// Error log format: YYYY/MM/DD HH:MM:SS [LEVEL] PID#TID: *CID MESSAGE
		print("Monitoring Nginx logfile at " + settings.Nginx.ErrorLogfilename)

		if _, err := os.Stat(settings.Nginx.ErrorLogfilename); err == nil {
			go monitorLog(settings.Nginx.ErrorLogfilename, loglines)

		} else if os.IsNotExist(err) {
			print("Nginx logfile does not exist")
		}
	}
}

func monitorLog(filename string, loglines chan LogLine) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {

		var logline LogLine
		logline.LogPath = filename

		var expression = regexp.MustCompile(`(?P<dateandtime>[(\d\/ \:]+) \[(?P<errorlevel>[a-z]+)\] (\d+)\#(\d+): \*?(\d+)? ?(?P<description>.*)`)
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

		fmt.Printf("Time: %s Error Level: %s Description: %s\n", logline.TimeStamp, logline.ErrorLevel, logline.Description)
	}

}

func StopReadingLogs() {
	stopLogMonitoring = true
}
