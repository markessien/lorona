package main

import (
	"encoding/json"
	"flag"
	"runtime/debug"
	"time"
)

var lLog *Logger

// This contains the result format that will be converted to json
// and sent to the remove receiver.
type Results struct {
	ContainerName        string
	ContainerSupport     string
	ContainerDescription string
	FileFormat           string
	SysMonitorInfo       SysMonitorInfo
	UptimeList           []UptimeResponse
	LoglineList          []LogLine
	BackupInfoList       []BackupInfo
	LogSummary           map[string]LogSummary
}

// Let's go!
func main() {

	// Get command line arguments
	settingsFilePtr := flag.String("settings", "settings.yaml", "Location of the settings file")
	logFilePtr := flag.String("log", "lorona.log", "Location of the log file")
	flag.Parse()

	// Configure logging
	logConfig := LogConfig{true, true, true, *logFilePtr, "lorona.log", 1024 * 1024, 10, 60 * 60 * 24 * 30}
	lLog = ConfigureLogging(logConfig)

	// Capture and log Panics
	defer func() {
		if r := recover(); r != nil {
			lLog.Print(string(debug.Stack()))
		}
	}()

	// Start!
	lLog.Print("Welcome to Lorona!")

	settings, err := LoadSettings(*settingsFilePtr)
	if err != nil {
		lLog.Fatal().Err(err).Msg("Could not load settings file")
	}

	lLog.Print("Lorona for package: " + settings.ContainerName + ". Settings File is " + *settingsFilePtr)

	process(settings)
}

// This function turns on all the monitoring threads
func process(settings *Settings) {

	// Create a channel queue that is as large as the number of threads we have.
	uptimes := make(chan UptimeResponse, len(settings.UptimeRequestList))
	loglines := make(chan LogLine, 100)
	sysinfos := make(chan SysMonitorInfo)
	backups := make(chan BackupInfo)

	// Reset the results structure and set the base things from the settings
	var results Results
	ResetResult(settings, &results)

	// Start all the monitoring services

	// Monitor specified endpoints to make sure they are up and running
	StartEndpointMonitoring(settings, uptimes)

	// Monitor the specified log files and send the log lines to this thread
	// for further processing
	StartLogMonitoring(settings, loglines)

	// Monitor the system - CPU, Ram and Diskspace on specified directories
	StartSystemMonitoring(settings, sysinfos)

	// Monitor backups
	StartBackupsMonitoring(settings, backups)

	go PromPublish()

	// Watch for messages from the channels and add them to the results structure
	// We need to handle the case that logs are filled faster than this function
	// clears the results. Use extra timer / channel for this
	for {

		select {
		case logline := <-loglines:
			// Add logline to the result. The logline contains enough info for it to later
			// know which particular log it came from
			var description = logline.Description

			if len(logline.Description) > 19 {
				logline.Description = logline.Description[0:20]
			}

			lLog.Print("Time: %s Error Level: %s Description: %s\n", logline.TimeStamp, logline.Severity, description)
			results.LoglineList = append(results.LoglineList, logline)
			UpdateMetrics(&results)
		case uptime := <-uptimes:

			// Add this new uptime result to the list
			results.UptimeList = append(results.UptimeList, uptime)
			// UpdateMetrics(&results)

		case sysinfo := <-sysinfos:
			results.SysMonitorInfo = sysinfo
			UpdateMetrics(&results)

		case backupInfo := <-backups:
			results.BackupInfoList = append(results.BackupInfoList, backupInfo)
			// UpdateMetrics(&results)

		case <-time.After(time.Second * 5): // does this do what we think it does? Check.
		default:

			s, _ := json.Marshal(results)
			lLog.Print(string(s))
			time.Sleep(5 * time.Second)

			ResetResult(settings, &results)
		}

	}

	StopEndpointMonitoring()
}

func ResetResult(settings *Settings, results *Results) {
	results.FileFormat = "LoronaV1"
	results.ContainerName = settings.ContainerName
	results.ContainerSupport = settings.ContainerSupport
	results.ContainerDescription = settings.ContainerDescription
	results.UptimeList = []UptimeResponse{}
	results.LoglineList = []LogLine{}
	results.BackupInfoList = []BackupInfo{}
	results.LogSummary = make(map[string]LogSummary)
}
