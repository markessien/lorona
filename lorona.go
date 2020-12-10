package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Welcome to Lorona!")

	settings, err := LoadSettings("settings.yaml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Lorona for package: " + settings.ContainerName)
}
