package main

import (
	"encoding/json"
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

	var results Results
	ResetResult(settings, &results)

	for {

		select {
		case uptime := <-uptimes:

			// Add this new uptime result to the list
			results.UptimeList = append(results.UptimeList, uptime)
		default:

			s, _ := json.Marshal(results)
			fmt.Println(string(s))
			time.Sleep(5 * time.Second)
			fmt.Println()
			fmt.Println()
			fmt.Println()

			ResetResult(settings, &results)
		}

	}

	StopEndpointMonitoring()
}

func ResetResult(settings *Settings, results *Results) {
	results.ContainerName = settings.ContainerName
	results.ContainerSupport = settings.ContainerSupport
	results.ContainerDescription = settings.ContainerDescription
	results.UptimeList = []UptimeResponse{}
	results.Loglines = []LogLine{}
}
