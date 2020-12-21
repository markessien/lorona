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
var loadAvg5 = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_system_load_avg_5", Help: "5 Minutues average system load"})
var loadAvg15 = promauto.NewGauge(prometheus.GaugeOpts{Name: "lorona_system_load_avg_15", Help: "15 Minutues average system load"})
var driveSpace = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_hd_available", Help: "The amount of space available in the hard drive"}, []string{"drive_paths"})

// Endpoint monitoring
var endpointAvailable = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_endpoint_up", Help: "1 or 0, depending on if the endpoint is up or not"}, []string{"urls"})
var endpointDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "lorona_endpoint_duration", Help: "Informs how long it took for the endpoint to respond"}, []string{"urls"})

// To be called by mainthread anytime there is something new to
// share with prometheus
func UpdateMetrics(result *Results) {

	// Publish system variables
	upTime.Set(float64(result.SysMonitorInfo.Uptime))
	cpuUsage.Set(float64(result.SysMonitorInfo.CpuUsagePercent))
	loadAvg1.Set(float64(result.SysMonitorInfo.LoadAveragePercent1))
	loadAvg5.Set(float64(result.SysMonitorInfo.LoadAveragePercent5))
	loadAvg15.Set(float64(result.SysMonitorInfo.LoadAveragePercent15))

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
}

func PromPublish() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
