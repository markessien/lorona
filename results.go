package main

import (
	"encoding/json"
	"fmt"
)

// Config struct for webapp config
type Results struct {
	ContainerName        string
	ContainerSupport     string
	ContainerDescription string
	FileFormat           string
	UptimeList           []UptimeResponse
	Loglines             []LogLine
}

func toString() {

	b := UptimeResponse{}
	m := Results{}
	m.UptimeList = []UptimeResponse{b}

	s, _ := json.Marshal(m)

	fmt.Printf("Got endpoint: " + string(s))

}
