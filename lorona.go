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

	process(settings)
}

func process(settings *Settings) {

	// Create a channel queue that is as large as the number of threads we have.
	uptimes := make(chan UptimeResponse, len(settings.UptimeRequestList))
	loglines := make(chan LogLine, 100)

	// Reset the results structure and set the base things from the settings
	var results Results
	ResetResult(settings, &results)

	// Start all the monitoring services
	// StartEndpointMonitoring(settings, uptimes)
	StartLogMonitoring(settings, loglines)

	// Watch for messages from the channels and add them to the results structure
	// We need to handle the case that logs are filled faster than this function
	// clears the results. Use extra timer / channel for this
	for {

		select {
		case logline := <-loglines:
			// Add logline to the result. The logline contains enough info for it to later
			// know which particular log it came from
			fmt.Printf("Time: %s Error Level: %s Description: %s\n", logline.TimeStamp, logline.ErrorLevel, logline.Description[0:20])
			results.Loglines = append(results.Loglines, logline)
		case uptime := <-uptimes:

			// Add this new uptime result to the list
			results.UptimeList = append(results.UptimeList, uptime)

		case <-time.After(time.Second * 5): // does this do what we think it does? Check.
		default:

			s, _ := json.Marshal(results)
			fmt.Println(string(s))
			time.Sleep(5 * time.Second)

			ResetResult(settings, &results)
		}

	}

	StopEndpointMonitoring()
}

func ResetResult(settings *Settings, results *Results) {
	results.FileFormat = "LoronaV1"
	results.ContainerName = settings.ContainerName
	results.ContainerSupport = settings.ContainerSupport
	results.ContainerDescription = settings.ContainerDescription
	results.UptimeList = []UptimeResponse{}
	results.Loglines = []LogLine{}
}
