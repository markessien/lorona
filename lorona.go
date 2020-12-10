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
	time.Sleep(16 * time.Second)
	shutdown()
}

func startup(settings *Settings) {
	StartEndpointMonitoring(settings)
}

// Request all threads to shutdown
func shutdown() {
	StopEndpointMonitoring()
}
