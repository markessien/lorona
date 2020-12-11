package main

import (
	"fmt"
	"log"
	"strconv"
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

	uptimes := make(chan UptimeResponse)

	StartEndpointMonitoring(settings, uptimes)

	for {

		uptime := <-uptimes

		fmt.Println("Received uptime info: " + uptime.Endpoint + " - Code:" + strconv.Itoa(uptime.ResponseCode))
		time.Sleep(5 * time.Second)

	}

	StopEndpointMonitoring()
}
