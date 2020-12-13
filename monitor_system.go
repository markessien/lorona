package main

import (
	"fmt"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// https://github.com/shirou/gopsutil
// https://github.com/ricochet2200/go-disk-usage/tree/master/du

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
