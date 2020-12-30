package main

import (
	"math"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Docs: https://godoc.org/github.com/prometheus/client_golang/prometheus#example-CounterVec
// Metric types: Counter (increases), Gauge (fluctuates), Histogram (Bucket Values), Summary

// System
var upTime = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_system_uptime", Help: "The uptime of this machine"})
var cpuUsage = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_system_cpu_usage", Help: "The current CPU usage of this machine"})
var loadAvg1 = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_system_load_avg_1", Help: "1 Minute average system load"})
var driveSpace = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_hd_available", Help: "The amount of space available in the hard drive"}, []string{"drive_path", "available", "growth_rate", "full_in", "physical_drive"})

// Endpoint monitoring
var endpointAvailable = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_endpoint_up", Help: "1 or 0, depending on if the endpoint is up or not"}, []string{"urls"})
var endpointDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_endpoint_duration", Help: "Informs how long it took for the endpoint to respond"}, []string{"urls"})

// Backups monitoring
var backupInfoGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_backup_info", Help: "1 or 0, depending on if a backup was done or not"}, []string{"backup_directory", "backup_in_last_24_hours", "last_backup_size", "last_backup_time", "last_backup_file"})

// Logs monitoring
var statusCodes = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_status_codes", Help: "A guage for each status_code, showing its count"}, []string{"log_path", "status_code"})
var severity = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_severity", Help: "A gauge for each severity, showing its count"}, []string{"log_path", "severity"})

// To be called by mainthread anytime there is something new to
// share with prometheus
func UpdateMetrics(result *Results) {

	// Publish system variables
	upTime.Set(float64(result.SysMonitorInfo.Uptime))
	cpuUsage.Set(float64(result.SysMonitorInfo.CpuUsagePercent))
	loadAvg1.Set(float64(result.SysMonitorInfo.LoadAveragePercent1))

	for _, driveUsage := range result.SysMonitorInfo.DriveUsage {
		// "drive_path", "available", "growth_rate", "full_in", "physical_drive"

		days := strconv.FormatFloat(driveUsage.DaysTillFull, 'f', 3, 64)

		if math.IsInf(driveUsage.DaysTillFull, 0) {
			days = "10 years"
		}

		driveSpace.WithLabelValues(driveUsage.Path,
			strconv.FormatFloat(driveUsage.PercentUsed, 'f', 3, 64),
			strconv.FormatUint(driveUsage.GrowthPerDayBytes, 10),
			days,
			driveUsage.VolumeName).Set(driveUsage.PercentUsed)
	}

	// Publish endpoints being monitored
	for _, uptimeResponse := range result.UptimeList {

		if uptimeResponse.ResponseCode == 200 {
			endpointAvailable.WithLabelValues(uptimeResponse.Endpoint).Set(1)
		} else {
			endpointAvailable.WithLabelValues(uptimeResponse.Endpoint).Set(0)
		}

		endpointDuration.WithLabelValues(uptimeResponse.Endpoint).Set(uptimeResponse.ResponseTime.Seconds())
	}

	for _, backupInfo := range result.BackupInfoList {

		/*
			if backupInfo.WasBackedUp {
				backupsDone.WithLabelValues(backupInfo.Folder).Set(1)
			} else {
				backupsDone.WithLabelValues(backupInfo.Folder).Set(0)
			}
		*/

		// {"backup_directory", "backup_in_last_24_hours", "last_backup_size", "last_backup_date", "last_backup_time"})

		// backupsSize.WithLabelValues(backupInfo.Folder).Set(float64(backupInfo.BackupFileSize))

		backupInfoGauge.WithLabelValues(backupInfo.Folder,
			btoa(backupInfo.WasBackedUp),
			itoa(backupInfo.LastBackupSize),
			ttoa(backupInfo.LastBackupTime),
			backupInfo.LastBackupFile).Set(btof(backupInfo.WasBackedUp))
	}

	// TODO: This loop is not needed, you can build the summary on the first loop
	var too_many_lines = 500
	for _, logLine := range result.LoglineList {

		summary, ok := result.LogSummary[logLine.LogPath]

		if ok == false {
			summary = LogSummary{}
			summary.StatusCount = make(map[string]int64)
			summary.SeverityLevelCount = make(map[string]int64)
		}

		summary.StatusCount[logLine.StatusCode] = summary.StatusCount[logLine.StatusCode] + 1

		if len(logLine.Severity) > 0 {
			summary.SeverityLevelCount[logLine.Severity] = summary.SeverityLevelCount[logLine.Severity] + 1
		}

		result.LogSummary[logLine.LogPath] = summary

		if too_many_lines <= 0 {
			// Pending a better solution, let's not allow the processing
			// of too many lines, to not kill the server
			print("Too many lines for a single tick to process")
			break
		}

	}

	// Set the values for the logs. We use two labels (logpath, code)
	for logFilePath, logSummary := range result.LogSummary {

		for s, value := range logSummary.StatusCount {
			statusCodes.WithLabelValues(logFilePath, s).Set(float64(value))
		}

		for s, value := range logSummary.SeverityLevelCount {
			severity.WithLabelValues(logFilePath, s).Set(float64(value))
		}

	}
}

func PromPublish() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
