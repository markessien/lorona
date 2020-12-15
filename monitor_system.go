package main

import (
	"fmt"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/v3/mem"
)

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
	AlertLevel  string   `yaml:"alert-level"`
	RepeatAlert string   `yaml:"repeat-alert"`
	Locations   []string `yaml:"locations"`
}

// The structure that saves the request for system info
// from the settings file
type SystemMonitorRequest struct {
	Cpu        SystemMonitorItem `yaml:"cpu"`
	Ram        SystemMonitorItem `yaml:"ram"`
	DriveSpace SystemMonitorItem `yaml:"drivespace"`
}

func StartSystemMonitoring(settings *Settings, sysinfos chan SystemMonitorRequest) error {

	stopSystemMonitoring = false

	go monitorSystem(sysinfos)

	return nil
}

func StopSystemMonitoring() {
	stopSystemMonitoring = true
}

// Keeps polling the system to get all the values and pushes them
// via a channel to the main thread
func monitorSystem(sysinfos chan SystemMonitorRequest) {

	// Docs are here: https://godoc.org/github.com/shirou/gopsutil

	v, _ := mem.VirtualMemory()

	u, _ := host.Uptime()

	fmt.Printf("Test: %v", u)
	// almost every return value is a struct
	fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

	// convert to JSON. String() is also implemented
	fmt.Println(v)

}
