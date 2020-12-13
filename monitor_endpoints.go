package main

import (
	"net/http"
	"time"
)

type UptimeRequest struct {
	Endpoint       string `yaml:"url"`
	StatusCheck    string `yaml:"status-check"`
	ExpectedStatus string `yaml:"expected-status"`
	CheckInterval  string `yaml:"check-interval"`
	GetTokenUrl    string `yaml:"get-token_url"`
}

// The UptimeResponse structure is used to record the results
// from a single uptime query
type UptimeResponse struct {
	Endpoint      string
	ResponseValue string
	ResponseCode  int
	ResponseTime  time.Duration
	PageTitle     string
}

// Used to stop the monitoring threads neatly
var stopEndpointMonitoring bool

// Kick off all the uptime monitoring threads
func StartEndpointMonitoring(settings *Settings, uptimes chan UptimeResponse) {

	stopEndpointMonitoring = false

	// Loop over all uptime requests
	for _, uptimeRequest := range settings.UptimeRequestList {

		// Check that there is a valid endpoint to check
		if len(uptimeRequest.Endpoint) > 0 {

			// Discover the check frequency
			duration, err := time.ParseDuration(uptimeRequest.CheckInterval)
			if err != nil {
				print("Could not monitor endpoint: " + uptimeRequest.Endpoint + " - Check duration")
			} else {

				// Start the go-routine that will do the monitoring
				go monitorEndpoint(uptimeRequest.Endpoint, duration, uptimes)
			}
		}

	}
}

// Request all threads to stop
func StopEndpointMonitoring() {
	stopEndpointMonitoring = true
}

// A go-routine that regularly checks if an endpoint is up
func monitorEndpoint(endpointUrl string, interval time.Duration, uptimes chan UptimeResponse) {

	for {
		// A single uptime object that will store this uptime check results
		var uptime UptimeResponse
		uptime.Endpoint = endpointUrl

		// Call the endpoint. Measure how long it takes
		start := time.Now()
		response, err := http.Head(uptime.Endpoint)
		elapsed := time.Since(start)

		if err != nil {
			// We use code 598 for an error like 'host not found'
			uptime.ResponseCode = 598
			uptime.ResponseTime = 0
			uptime.ResponseValue = err.Error()
		} else {
			uptime.ResponseCode = response.StatusCode
			uptime.ResponseTime = elapsed
		}

		if stopEndpointMonitoring == true {
			return
		}

		// push it into the channel. It will be picked up by the main thread
		uptimes <- uptime

		// Sleep for the time from the settings file before testing uptime again
		time.Sleep(interval)
	}
}
