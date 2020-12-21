package main

import (
	"net/http"

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
var driveSpace = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_hd_available", Help: "The amount of space available in the hard drive"}, []string{"drive_paths"})

// Endpoint monitoring
var endpointAvailable = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_endpoint_up", Help: "1 or 0, depending on if the endpoint is up or not"}, []string{"urls"})
var endpointDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_endpoint_duration", Help: "Informs how long it took for the endpoint to respond"}, []string{"urls"})

// Backups monitoring
var backupsDone = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_backup_done", Help: "1 or 0, depending on if a backup was done or not"}, []string{"directory"})
var backupsSize = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_backupsize", Help: "The size of the backup file"}, []string{"directory"})

// Logs monitoring
// https://lincolnloop.com/blog/tracking-application-response-time-nginx/
// - A graph showing each type of logline, stacked. E.g 200, 404 and so on
// We have a loglines which are all categorized
// - A graph of only Error 500
// - A graph of slow queries
// - A graph of avg upstream response time over tick interval
// - A graph of response time over tick interval

// Visualisation would ideally be like this: https://prometheus.io/docs/visualization/grafana/

// We return the number of errors since the last tick. Each lorona tick (of the log) provides a fixed number
// of log entries. These remain the values till the next tick. Every prometheus query will return this.

// Error 500 over what time?
// metrics: lorona_status_5xx (count)
// metrics: lorona_status_4xx
// metrics: lorona_status_2xx
// metrics: lorona_status_3xx
// metrics: lorona_status_other
// metrics: lorona_log_errors
// metrics: lorona_log_warnings

// To be called by mainthread anytime there is something new to
// share with prometheus
func UpdateMetrics(result *Results) {

	// Publish system variables
	upTime.Set(float64(result.SysMonitorInfo.Uptime))
	cpuUsage.Set(float64(result.SysMonitorInfo.CpuUsagePercent))
	loadAvg1.Set(float64(result.SysMonitorInfo.LoadAveragePercent1))

	for _, driveUsage := range result.SysMonitorInfo.DriveUsage {
		driveSpace.WithLabelValues(driveUsage.Path).Set(driveUsage.PercentUsed)
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

		if backupInfo.WasBackedUp {
			backupsDone.WithLabelValues(backupInfo.Folder).Set(1)
		} else {
			backupsDone.WithLabelValues(backupInfo.Folder).Set(0)
		}

		backupsSize.WithLabelValues(backupInfo.Folder).Set(float64(backupInfo.BackupFileSize))
	}

	var too_many_lines = 500
	for _, logLine := range result.LoglineList {

		summary, ok := result.LogSummary[logLine.LogPath]

		if ok == false {
			summary = LogSummary{}
			summary.StatusCount = make(map[string]int64)
			summary.ErrorLevelCount = make(map[string]int64)
		}

		summary.StatusCount[logLine.StatusCode] = summary.StatusCount[logLine.StatusCode] + 1
		summary.ErrorLevelCount[logLine.ErrorLevel] = summary.ErrorLevelCount[logLine.ErrorLevel] + 1

		result.LogSummary[logLine.LogPath] = summary

		if too_many_lines <= 0 {
			// Pending a better solution, let's not allow the processing
			// of too many lines, to not kill the server
			print("Too many lines for a single tick to process")
			break
		}

	}

}

func PromPublish() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
