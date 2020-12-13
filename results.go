package main

// Config struct for webapp config
type Results struct {
	ContainerName        string
	ContainerSupport     string
	ContainerDescription string
	FileFormat           string
	UptimeList           []UptimeResponse
	Loglines             []LogLine
}
