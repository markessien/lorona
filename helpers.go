package main

import (
	"fmt"
	"strconv"
	"time"
)

func btof(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func btoa(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func ttoa(t time.Time) string {
	return t.Format(time.RFC850)
}

// Print ersatz
func print(str string) {
	fmt.Println(str)
}
