package main

import (
	"time"

	"github.com/shirou/gopsutil/v3/host"
)

// Used to stop the monitoring threads neatly
var stopSystemMonitoring bool

// A structure for providing info about the usage of a single drive
type SysDriveInfo struct {
	Path        string
	PercentUsed float64
	Capacity    uint64
	Used        uint64
}

// The structure that saves all the information about the
// system that we want to provide
type SysMonitorInfo struct {
	HostName           string
	NetAddress         string
	CpuUsagePercent    float64
	SystemWarnings     []string
	LoadAveragePercent float64
	DriveUsage         []SysDriveInfo
}

// A structure that represents a single request for system
// info - e.g just CPU
type SystemMonitorItem struct {
	AlertLevel  string
	RepeatAlert string
	Locations   []string
}

// The structure that saves the request for system info
// from the settings file
type SystemMonitorRequest struct {
	Cpu           SystemMonitorItem `yaml:"cpu"`
	Ram           SystemMonitorItem `yaml:"ram"`
	DriveSpace    SystemMonitorItem `yaml:"drivespace"`
	CheckInterval string            `yaml:"check-interval"`
}

func StartSystemMonitoring(settings *Settings, sysinfos chan SysMonitorInfo) error {

	stopSystemMonitoring = false

	duration, err := time.ParseDuration(settings.SysMonitorRequest.CheckInterval)
	if err != nil {
		print("Could not parse requested interval: " + settings.SysMonitorRequest.CheckInterval)
	} else {
		go monitorSystem(settings.SysMonitorRequest, duration, sysinfos)
	}

	return nil
}

func StopSystemMonitoring() {
	stopSystemMonitoring = true
}

// https://github.com/shirou/gopsutil
// https://github.com/ricochet2200/go-disk-usage/tree/master/du

// Interested in:
// CPU Percent: https://godoc.org/github.com/shirou/gopsutil/cpu (ALL CPUs and Combined)
// Disk Usage: https://godoc.org/github.com/shirou/gopsutil/disk#Usage
// HostID: https://godoc.org/github.com/shirou/gopsutil/host#HostID
// Warnings: https://godoc.org/github.com/shirou/gopsutil/host#Warnings
// Load Avg: https://godoc.org/github.com/shirou/gopsutil/load#Avg
// Net Addre: https://godoc.org/github.com/shirou/gopsutil/net#Addr
// Docker per item: https://godoc.org/github.com/shirou/gopsutil/net#Addr

// Because of volume of info, limit it

// Keeps polling the system to get all the values and pushes them
// via a channel to the main thread
func monitorSystem(request SystemMonitorRequest, interval time.Duration, sysinfos chan SysMonitorInfo) {

	// Docs are here: https://godoc.org/github.com/shirou/gopsutil

	sys := SysMonitorInfo{}
	sys.HostName, _ = host.HostID()

	/*
		v, _ := mem.VirtualMemory()

		u, _ :=

			fmt.Printf("Test: %v", u)
		// almost every return value is a struct
		fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

		// convert to JSON. String() is also implemented
		fmt.Println(v)
	*/

	for {
		time.Sleep(interval)

		sysinfos <- sys
	}

}
