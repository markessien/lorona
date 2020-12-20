package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Docs: https://godoc.org/github.com/prometheus/client_golang/prometheus#example-CounterVec
// Metric types: Counter (increases), Gauge (fluctuates), Histogram (Bucket Values), Summary
// What do we want to expose?
// Uptime: Counter
// CpuUsage: Gauge
// LoadAverage1, 5, 15: Gauge
// Drive1: PercentageUsed - Gauge
// Drive1: Capacity - Gauge
// Drive1: Used - Gauge
// Drive2: PercentageUsed - Gauge
// Drive2: ...
//
// EndPoint1: Up? Gauge
// EndPoint1: Responsetime - Gauge

var upTime = promauto.NewGauge(prometheus.GaugeOpts{Name: "Uptime", Help: "The uptime of this machine"})

// To be called by mainthread anytime there is something new to
// share with prometheus
func UpdateMetrics(result *Results) {
	upTime.Set(float64(result.SysMonitorInfo.Uptime))
}

func PromPublish() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
