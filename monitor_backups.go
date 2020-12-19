package main

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

// When we observe a file, we serve it in this persistent
// variable. We can use this to discover new files added
var knownBackupFiles []string
var tickRepeatFrequency time.Duration

// The UptimeResponse structure is used to record the results
// from a single uptime query
type BackupInfo struct {
	Folder              string
	LastBackup          time.Time
	WasBackedUp         bool
	ErrorMessage        string
	ExistingBackupFiles []string
}

type BackupMonitorRequest struct {
	Folder               string `yaml:"backup-folder"`
	MinimumFileSize      string `yaml:"backup-minimum-file-size"`
	CheckBackupFrequency string `yaml:"check-backups-frequency"`
	CheckBackupTime      string `yaml:"check-backups-time"`
	RemoteBackupUrl      string `yaml:"remote-backup-url"`
	RemoteBackupFormat   string `yaml:"remote-backup-format"`
}

func StartBackupsMonitoring(settings *Settings, backups chan BackupInfo) error {

	dataFile, err := os.Open("./known_backup_files.dat")
	if err == nil {
		dataDecoder := gob.NewDecoder(dataFile)
		err = dataDecoder.Decode(&knownBackupFiles)
	}
	dataFile.Close()

	// We use timers here. Each time they tick, we check how the backup situation
	// is currently looking. The timer ticks run in separate threads
	for i, backupMonitorRequest := range settings.BackupMonitorRequest {

		// Parse the time. The format is fixed in 24 hour time format (the 15:04 shows that)
		t, err := time.Parse("15:04", backupMonitorRequest.CheckBackupTime)

		// Next, we apply that time to todays date, so we know when it's going to tick today
		// Get a datetime representing when the first check would happen today
		// This is based on the requested backup check time
		firstTick := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

		// Get the time the tick after the first would happen
		nextTick := firstTick.Add(time.Hour * 24)

		// Select if the very first tick will be the tick today or tommorow
		startTick := nextTick
		if firstTick.After(time.Now()) {
			// This means we will do a check today
			startTick = firstTick
		}

		// Get duration till our next tick
		timeTillStart := time.Until(startTick)
		timeTillStart = 1000 * 1000 * 3000
		tickRepeatFrequency = timeTillStart
		fmt.Println(timeTillStart)

		if err == nil {
			if len(settings.BackupMonitorRequest) > 0 {
				f := startBackupMonitoring(backupMonitorRequest.Folder, backups)

				// Let's shift each loop by 5seconds so the checks do not all run exactly in parallel
				s := strconv.Itoa(i*5) + "s"
				shift, _ := time.ParseDuration(s)
				_ = time.AfterFunc(timeTillStart+shift, f)
			}
		} else {
			print("Was not able to parse check time for " + backupMonitorRequest.Folder)
		}
	}

	return nil
}

func startBackupMonitoring(backupFolder string, backups chan BackupInfo) func() {
	return func() {
		checkForBackups(backupFolder, backups)
	}
}

func checkForBackups(backupFolder string, backups chan BackupInfo) {

	// Let's create the next tick first of all
	// f := startBackupMonitoring(backupFolder, backups)
	// _ = time.AfterFunc(tickRepeatFrequency, f)
	// ---

	var backupInfo BackupInfo

	// Open the backup folder to list the files in it
	files, err := ioutil.ReadDir(backupFolder)
	if err != nil {

		// If it fails, then obviously no backup
		backupInfo.Folder = backupFolder
		backupInfo.WasBackedUp = false
		backupInfo.LastBackup = time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC)
		backupInfo.ErrorMessage = "Backup folder not found"
		backups <- backupInfo
		return
	}

	if len(files) > 300 {
		backupInfo.Folder = backupFolder
		backupInfo.WasBackedUp = false
		backupInfo.LastBackup = time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC)
		backupInfo.ErrorMessage = "Too many files in the backup folder"
		backups <- backupInfo
		return
	}

	if len(files) == 0 {
		backupInfo.Folder = backupFolder
		backupInfo.WasBackedUp = false
		backupInfo.LastBackup = time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC)
		backupInfo.ErrorMessage = "No files in the backup folder"
		backups <- backupInfo
		return
	}

	var existingBackupFiles []string

	// List all files in the backup folder
	for _, f := range files {

		fileStat, err := os.Stat(f.Name())

		if err == nil && fileStat.Size() > 1024*1024*30 {
			print("Backup found")
		}

		/*
					existingBackupFiles = append(existingBackupFiles, f.Filepath)
			found := false
			for _, b := range knownBackupFiles {
				if f.Filepath == b {
					// We already know this backup file
					found = true
					break
				}
			}

			backupFound = false
			if found == false {
				// We have a new file in the directory
				fileStat := os.Stat(f.Filepath)

				if fileStat.Size() > 30*MB {
					backupFound = true
				}
			}
		*/
	}

	backupInfo.ExistingBackupFiles = existingBackupFiles
	backups <- backupInfo

}
