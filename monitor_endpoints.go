package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var stopEndpointMonitoring bool

func StartEndpointMonitoring(settings *Settings) {

	// Create channel so we can use it to stop the processes
	stopEndpointMonitoring = false

	// Check if we need general endpoint monitoring
	if len(settings.Uptime.Endpoint) > 0 {
		duration, err := time.ParseDuration(settings.Uptime.CheckInterval)
		if err != nil {
			log.Fatal(err)
		}

		go monitorGeneralEndpoint(settings.Uptime.Endpoint, duration)
	}

}

func StopEndpointMonitoring() {
	stopEndpointMonitoring = true
}

func monitorGeneralEndpoint(endpointUrl string, interval time.Duration) {

	fmt.Printf("Entered monitoring routine")

	fmt.Printf("Got endpoint: " + endpointUrl)
	for {
		response, err := http.Head(endpointUrl)
		if err != nil {
			log.Fatal("Error: Unable to download URL (", endpointUrl, ") with error: ", err)
		}

		fmt.Printf("Requested " + strconv.Itoa(response.StatusCode))

		if response.StatusCode != http.StatusOK {
			log.Fatal("Error: HTTP Status = ", response.Status)
		}

		time.Sleep(interval)

		if stopEndpointMonitoring == true {
			fmt.Printf("Stopping General Monitoring")
			return
		}
	}
}
