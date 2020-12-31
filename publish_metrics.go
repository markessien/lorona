package main

import (
	"math"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// System
var upTimeGauge = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_system_uptime", Help: "The uptime of this machine"})
var cpuUsageGauge = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_system_cpu_usage", Help: "The current CPU usage of this machine"})

// Memory
var memUsagePercentGauge = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_mem_usage_percent", Help: "The current CPU usage of this machine"})
var memTotalGauge = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_mem_total", Help: "The current CPU usage of this machine"})
var memAvailableGauge = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_mem_available", Help: "The current CPU usage of this machine"})

// Bandwidth
var bandwidthUsageTotalGauge = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_bandwidth_usage", Help: "The current CPU usage of this machine"})
var bandwidthUsageSentGauge = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_bandwidth_sent_usage", Help: "The current CPU usage of this machine"})
var bandwidthUsageRecvGauge = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_bandwith_recv_usage", Help: "The current CPU usage of this machine"})

// Hard drive space
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
	upTimeGauge.Set(float64(result.SysMonitorInfo.Uptime))
	cpuUsageGauge.Set(float64(result.SysMonitorInfo.CpuUsagePercent))

	// Memory
	memUsagePercentGauge.Set(result.SysMonitorInfo.MemUsagePercent)
	memTotalGauge.Set(float64(result.SysMonitorInfo.MemTotal))
	memAvailableGauge.Set(float64(result.SysMonitorInfo.MemAvailable))

	// Bandwidth
	bandwidthUsageTotalGauge.Set(float64(result.SysMonitorInfo.BandwidthUsageTotal))
	bandwidthUsageSentGauge.Set(float64(result.SysMonitorInfo.BandwidthUsageSent))
	bandwidthUsageRecvGauge.Set(float64(result.SysMonitorInfo.BandwidthUsageRecv))

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
			lLog.Print("Too many lines for a single tick to process")
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
