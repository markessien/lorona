package main

import (
	"math"
	"path/filepath"
	"time"

	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
)

// TODO: Add physical drive name

// Used to stop the monitoring threads neatly
var stopSystemMonitoring bool

// For measuring bandwidth
var lastNetworkBytesSent uint64
var lastNetworkBytesRecv uint64
var lastNetworkTimeStamp time.Time

// A structure for providing info about the usage of a single drive
type SysDriveInfo struct {
	Path               string
	PercentUsed        float64
	Capacity           uint64
	Used               uint64
	Fstype             string
	VolumeName         string
	DaysTillFull       float64
	GrowthPerDayBytes  uint64
	UsedCheckpoint     uint64 // Used to measure growth rate
	UsedCheckpointTime time.Time
}

// The structure that saves all the information about the
// system that we want to provide
type SysMonitorInfo struct {
	HostName            string
	CpuUsagePercent     float64
	MemUsagePercent     float64
	MemTotal            uint64
	MemAvailable        uint64
	BandwidthUsageTotal uint64
	BandwidthUsageSent  uint64
	BandwidthUsageRecv  uint64
	SystemWarnings      []string // Can we get this?
	Uptime              uint64   // In seconds
	DriveUsage          map[string]*SysDriveInfo
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

	netw, _ := net.IOCounters(false) // for all networks aggregated (false)
	netw_all := netw[0]

	lastNetworkBytesSent = netw_all.BytesSent
	lastNetworkBytesRecv = netw_all.BytesRecv
	lastNetworkTimeStamp = time.Now()

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
	sys.DriveUsage = make(map[string]*SysDriveInfo)

	// Get Hostname
	// HostID: https://godoc.org/github.com/shirou/gopsutil/host#HostID
	sys.HostName, _ = host.HostID()

	for {

		sys.Uptime, _ = host.Uptime()

		// Get CPU usage since last call
		cpus, _ := cpu.Percent(0, false)
		sys.CpuUsagePercent = cpus[0]

		memory, _ := mem.VirtualMemory()
		sys.MemUsagePercent = memory.UsedPercent
		sys.MemTotal = memory.Total
		sys.MemAvailable = memory.Available

		netw, _ := net.IOCounters(false) // for all networks aggregated (false)
		netw_all := netw[0]

		curNetworkBytesSent := netw_all.BytesSent
		curNetworkBytesRecv := netw_all.BytesRecv

		timeElapsed := time.Now().Sub(lastNetworkTimeStamp)

		if timeElapsed.Seconds() > 0 {
			sys.BandwidthUsageSent = uint64(float64(curNetworkBytesSent-lastNetworkBytesSent) / timeElapsed.Seconds())
			sys.BandwidthUsageRecv = uint64(float64(curNetworkBytesRecv-lastNetworkBytesRecv) / timeElapsed.Seconds())
			sys.BandwidthUsageTotal = sys.BandwidthUsageSent + sys.BandwidthUsageRecv

			lastNetworkBytesSent = netw_all.BytesSent
			lastNetworkBytesRecv = netw_all.BytesRecv
			lastNetworkTimeStamp = time.Now()
		}

		// Loop over all the requests to monitor drives
		for _, drivePath := range request.DriveSpace.Locations {

			// var driveInfoResponse SysDriveInfo
			stat, err := disk.Usage(drivePath)
			if err == nil {

				_, ok := sys.DriveUsage[drivePath]
				if !ok {
					sys.DriveUsage[drivePath] = new(SysDriveInfo)
					sys.DriveUsage[drivePath].UsedCheckpointTime = time.Now()
					sys.DriveUsage[drivePath].UsedCheckpoint = stat.Used
				}

				sys.DriveUsage[drivePath].Path = stat.Path
				sys.DriveUsage[drivePath].Fstype = stat.Fstype
				sys.DriveUsage[drivePath].Used = stat.Used
				sys.DriveUsage[drivePath].Capacity = stat.Total
				sys.DriveUsage[drivePath].PercentUsed = stat.UsedPercent
				sys.DriveUsage[drivePath].VolumeName = filepath.VolumeName(drivePath)

				// The below figures out how fast the disk is filling up, and estimates
				// by when the disk will be full based on this growth rate

				measurementInterval := 1.0
				timeInterval := time.Now().Sub(sys.DriveUsage[drivePath].UsedCheckpointTime).Hours()
				if timeInterval > measurementInterval {

					spaceAddedBytes := sys.DriveUsage[drivePath].Used - sys.DriveUsage[drivePath].UsedCheckpoint
					intervalsInDay := 24 / timeInterval

					sys.DriveUsage[drivePath].GrowthPerDayBytes = uint64(math.Round(intervalsInDay * float64(spaceAddedBytes)))

					availableSpace := sys.DriveUsage[drivePath].Capacity - sys.DriveUsage[drivePath].Used
					sys.DriveUsage[drivePath].DaysTillFull = float64(availableSpace) / float64(sys.DriveUsage[drivePath].GrowthPerDayBytes)

					sys.DriveUsage[drivePath].UsedCheckpoint = sys.DriveUsage[drivePath].Used
					sys.DriveUsage[drivePath].UsedCheckpointTime = time.Now()

				}
			} else {
				// Reset the drive usage, so it does not infinitely grow
				if sys.DriveUsage[drivePath] != nil {
					sys.DriveUsage[drivePath].Fstype = stat.Fstype
					sys.DriveUsage[drivePath].Used = 0
					sys.DriveUsage[drivePath].Capacity = 0
					sys.DriveUsage[drivePath].PercentUsed = 0
				}
			}

		}

		// Push results to the main thread thorugh a channel
		sysinfos <- sys

		time.Sleep(interval)
	}

}
