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

	Uptime   []UptimeResponse
	Loglines []LogLine
}

func toString() {

	b := UptimeResponse{}
	m := Results{}
	m.Uptime = []UptimeResponse{b}

	s, _ := json.Marshal(m)

	fmt.Printf("Got endpoint: " + string(s))

}
