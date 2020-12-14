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

// Docs are here: https://godoc.org/github.com/shirou/gopsutil
func StartSystemMonitoring(settings *Settings) {
	v, _ := mem.VirtualMemory()

	u, _ := host.Uptime()

	fmt.Printf("Test: %v", u)
	// almost every return value is a struct
	fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

	// convert to JSON. String() is also implemented
	fmt.Println(v)

}
