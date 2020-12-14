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
	LastTimestamp     string
	LastByteRead      int64
	Regex             string
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

		// Start the go-routine that will be monitoring the logs
		go monitorLog(logFile, loglines)
	}
}

func StopReadingLogs() {
	stopLogMonitoring = true
}

func getFileLocalData() {

}

// Problems
// - Very large files - solved
// - Files rotated
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
	// print(strconv.Itoa(new_offset))

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

func findLastReadSpot() {
	// Complicated technique. This is the way:
	// We use the timestamp to mark our position. That means for
	// each log file, we store some info locally on last read
	// time stamp. If there is no info stored, we only read up to 24 hours prior

	// We also store the byte offset where we last read from. This is just a hint
	// as the file may be deleted or rotated.

	// Step 1: Get last open byte offset.
	// Step 2: Retrieve our last gotten timestamp for this log file
	// Step 3: Open file at last byte offset. Walk backwards till we find our timestamp. Start reading from there. Stop at 24 hours earlier at most
	// ALGA: If we nothing stored, we start from end and walk backwards till we find a log entry older than 24 hours ago. We start there.
	// If last byte read is larger than file, we cancel it and use ALGA
	//

	// ### The block below is to seek to the last opened position for this file
	// Purpose of this is to avoid us loading 60GB log files into memory and crashing
	// The system. The log parser is very conservative with memory usage and will always
	// only scan from the end of the file

	/*
		last_10_log_lines = []

		// Note that the log file may have been rotated
		var last_pos = 0
		o2, err := f.Seek(last_pos, 0)

		var last_timestamp = ""

		if len(last_timestamp) > 0 {
			cur_pos = last_pos
			eof_count = 0
			// Walk backwards till we find the last known timestamp
			for i := 0; i < 100; i++ {
				// We are expecting the timestamp to be very close to the offset, but we limit our
				// backwards search to 100 * 1000. The longest entry in our sample log is 644 chars,
				// so we assume 1000 chars for a very long log entry
				// So in total, we will be walking 100kb backwards. If we find 3 newlines, we will
				// stop the search, as it means something is wrong

				// Second Parameter is the point of reference for offset
				// 0 = Beginning of file
				// 1 = Current position
				// 2 = End of file
				// We jump back 1000 bytes and then search forward for either our timestamp or an EOF
				o2, err := f.Seek(-1000, 1)
				cur_pos = cur_pos - 1000

				// Read 1000 bytes
				b1 := make([]byte, 1000)
				n1, err := f.Read(b1)

				// Convert the bytes to a string
				readline = string(b1[:n1])

				// Check if the timestamp is in this block. If so, then
				// seek to that timestamp position and exit
				timestamp_pos = strings.Index(readline, last_timestamp)
				if timestamp_pos > 0 {
					o2, err := f.Seek(cur_pos+timestamp_pos, 0)
					break
				}

				newline_pos = strings.Index(readline, "\n")
				if newline_pos > 0 {
					eof_count++
					if eof_count > 20 break
				}
			}
		}

		// Open with scanner so we can check each line
		// scanner := bufio.NewScanner(f)

		// might make sense to only read 10MB from the end

		/*
				// We can use the below to jump to where we want
				pos := start
			    scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			        advance, token, err = bufio.ScanLines(data, atEOF)
			        pos += int64(advance)
			        return
				}
				scanner.Split(scanLines)
	*/

	/*
		b1 := make([]byte, 5)
		n1, err := f.Read(b1)
	*/

}
