package main

// The UptimeResponse structure is used to record the results
// from a single uptime query
type BackupInfo struct {
	Folder      string
	WasBackedUp bool
}

type BackupMonitorRequest struct {
	Folder               string `yaml:"backup-folder"`
	MinimumFileSize      string `yaml:"backup-minimum-file-size"`
	CheckBackupFrequency string `yaml:"check-backups"`
	CheckBackupTime      string `yaml:"check-backups-time"`
	RemoteBackupUrl      string `yaml:"remote-backup-url"`
	RemoteBackupFormat   string `yaml:"remote-backup-format"`
}

func StartBackupsMonitoring(settings *Settings, backups chan BackupInfo) error {

	if len(settings.BackupMonitorRequest) > 0 {
		go monitorBackups(settings, backups)
	}

	return nil
}

func monitorBackups(settings *Settings, backups chan BackupInfo) {

	for _, backupMonitorRequest := range settings.BackupMonitorRequest {
		print(backupMonitorRequest.Folder)
	}
}
