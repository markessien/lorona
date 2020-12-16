package main

import (
	"time"

	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
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
	Fstype      string
}

// The structure that saves all the information about the
// system that we want to provide
type SysMonitorInfo struct {
	HostName             string
	CpuUsagePercent      float64
	SystemWarnings       []string // Can we get this?
	LoadAveragePercent1  float64
	LoadAveragePercent5  float64
	LoadAveragePercent15 float64
	Uptime               uint64
	DriveUsage           []SysDriveInfo
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

// Keeps polling the system to get all the values and pushes them
// via a channel to the main thread
func monitorSystem(request SystemMonitorRequest, interval time.Duration, sysinfos chan SysMonitorInfo) {

	sys := SysMonitorInfo{}

	// Get Hostname
	// HostID: https://godoc.org/github.com/shirou/gopsutil/host#HostID
	sys.HostName, _ = host.HostID()

	for {

		sys.Uptime, _ = host.Uptime()

		// Get CPU usage since last call
		cpus, _ := cpu.Percent(0, false)
		sys.CpuUsagePercent = cpus[0]

		// Get the average CPU load
		avg, _ := load.Avg()
		sys.LoadAveragePercent1 = avg.Load1
		sys.LoadAveragePercent5 = avg.Load5
		sys.LoadAveragePercent15 = avg.Load15

		// Loop over all the requests to monitor drives
		for _, drivePath := range request.DriveSpace.Locations {

			var driveInfoResponse SysDriveInfo
			stat, err := disk.Usage(drivePath)
			if err == nil {
				driveInfoResponse.Path = stat.Path
				driveInfoResponse.Fstype = stat.Fstype
				driveInfoResponse.Used = stat.Used
				driveInfoResponse.Capacity = stat.Total
				driveInfoResponse.PercentUsed = stat.UsedPercent
				sys.DriveUsage = append(sys.DriveUsage, driveInfoResponse)
			}

		}

		// Push results to the main thread thorugh a channel
		sysinfos <- sys

		// Reset the drive usage, so it does not infinitely grow
		sys.DriveUsage = []SysDriveInfo{}
		time.Sleep(interval)
	}

}
