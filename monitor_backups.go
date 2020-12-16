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

	// WE use timers here. Each time they tick, we check how the backup situation
	// is currently looking. The timer ticks run in separate threads
	for _, backupMonitorRequest := range settings.BackupMonitorRequest {

		// Parse the time. The format is fixed in 24 hour time format (the 15:04 shows that)
		t, err := time.Parse("15:04", backupMonitorRequest.CheckBackupTime)

		fmt.Println(t.Format(time.RFC850))
		if err == nil {
			if len(settings.BackupMonitorRequest) > 0 {
				f := startBackupMonitoring(backupMonitorRequest.Folder)
				_ = time.AfterFunc(1*time.Second, f)
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

func monitorBackups(settings *Settings, backups chan BackupInfo) {

}
