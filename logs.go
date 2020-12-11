package main

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

type Log struct {
	filepath      string
	LastTimestamp string
}

func StartReadingLogs() {

}

func StopReadingLogs() {

}
