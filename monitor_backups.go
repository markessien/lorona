package main

import (
	"fmt"
	"time"
)

// The UptimeResponse structure is used to record the results
// from a single uptime query
type BackupInfo struct {
	Folder      string
	WasBackedUp bool
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

	// We use timers here. Each time they tick, we check how the backup situation
	// is currently looking. The timer ticks run in separate threads
	for _, backupMonitorRequest := range settings.BackupMonitorRequest {

		// Parse the time. The format is fixed in 24 hour time format (the 15:04 shows that)
		t, err := time.Parse("15:04", backupMonitorRequest.CheckBackupTime)

		// Get a datetime representing when the first check would happen today
		// This is based on the requested backup check time
		firstTick := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

		// Get the time the next tick would happen
		nextTick := firstTick.Add(time.Hour * 24)

		// Select if the first tick will be the tick today or tommorow
		startTick := nextTick
		if firstTick.After(time.Now()) {
			// This means we will do a check today
			startTick = firstTick
		}

		fmt.Println(firstTick.Format(time.RFC850))
		fmt.Println(nextTick.Format(time.RFC850))
		fmt.Println(startTick.Format(time.RFC850))

		timeTillStart := time.Until(startTick)
		fmt.Println(timeTillStart)

		if err == nil {
			if len(settings.BackupMonitorRequest) > 0 {
				f := startBackupMonitoring(backupMonitorRequest.Folder)
				_ = time.AfterFunc(timeTillStart, f)
			}
		}

	}

	return nil
}

func startBackupMonitoring(backupFolder string) func() {
	return func() {
		checkForBackups(backupFolder)
	}
}

func checkForBackups(backupFolder string) {
	fmt.Printf("creating func with '%s'\n", backupFolder)
}
