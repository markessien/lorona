package main

import (
	"log"
	"net/http"
	"time"
)

type UptimeResponse struct {
	Endpoint      string
	ResponseValue string
	ResponseCode  int
	ResponseTime  time.Duration
	PageTitle     string
}

var stopEndpointMonitoring bool

// This is called to start the endpoint monitoring. The results
func StartEndpointMonitoring(settings *Settings, uptimes chan UptimeResponse) {

	stopEndpointMonitoring = false

	for _, uptimeRequest := range settings.UptimeRequestList {

		if len(uptimeRequest.Endpoint) > 0 {
			duration, err := time.ParseDuration(uptimeRequest.CheckInterval)
			if err != nil {
				print("Could not monitor endpoint: " + uptimeRequest.Endpoint + " - Check duration")
				log.Fatal(err)
			} else {
				go monitorEndpoint(uptimeRequest.Endpoint, duration, uptimes)
			}
		}

	}
}

func StopEndpointMonitoring() {
	stopEndpointMonitoring = true
}

func monitorEndpoint(endpointUrl string, interval time.Duration, uptimes chan UptimeResponse) {

	for {
		var uptime UptimeResponse
		uptime.Endpoint = endpointUrl

		// Call the endpoint
		start := time.Now()
		response, err := http.Head(uptime.Endpoint)
		elapsed := time.Since(start)

		if err != nil {
			uptime.ResponseCode = 598
			uptime.ResponseTime = 0
		} else {
			uptime.ResponseCode = response.StatusCode
			uptime.ResponseTime = elapsed
		}

		if stopEndpointMonitoring == true {
			return
		}

		uptimes <- uptime
		time.Sleep(interval)
	}
}
