package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("Welcome to Lorona!")

	settings, err := LoadSettings("settings.yaml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Lorona for package: " + settings.ContainerName)

	startup(settings)
}

func startup(settings *Settings) {

	// Create a channel queue that is as large as the number of threads we have.
	uptimes := make(chan UptimeResponse, len(settings.UptimeRequestList))

	StartEndpointMonitoring(settings, uptimes)

	for {

		select {
		case uptime := <-uptimes:

			fmt.Printf("Received uptime info for %s. Result Code: %d. RTT=%s. ", uptime.Endpoint, uptime.ResponseCode, uptime.ResponseTime)
			if uptime.ResponseCode == 598 {
				fmt.Printf(uptime.ResponseValue)
			}
			fmt.Println()

		default:

			fmt.Println("no activity")
			time.Sleep(5 * time.Second)

		}

	}

	StopEndpointMonitoring()
}
