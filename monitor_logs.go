package main

import "os"

type LogLine struct {
	AppName          string
	LogPath          string
	ErrorDescription string
	ErrorLevel       string
	TimeStamp        int64
	SourceIP         string
	Request          string
	Code             string
	UserAgent        string
	ResponseSize     int64
	Referrer         string
	Upstream         string
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
		print("Monitoring Nginx logfile at " + settings.Nginx.ErrorLogfilename)

		if _, err := os.Stat(settings.Nginx.ErrorLogfilename); err == nil {
			go monitorLog(loglines)

		} else if os.IsNotExist(err) {
			print("Nginx logfile does not exist")
		}
	}
}

func monitorLog(loglines chan LogLine) {

}

func StopReadingLogs() {
	stopLogMonitoring = true
}
