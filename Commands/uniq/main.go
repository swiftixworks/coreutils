package main

import "fmt"
import "os"
import "swiftix/userland"

func splitLines(data string) []string {
	lines := []string{}
	start := 0
	for index := 0; index < len(data); index++ {
		if data[index] == 10 {
			lines = append(lines, data[start:index])
			start = index + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

func main() {
	data, status := userland.ReadInput("uniq", os.Args[1:])
	lines := splitLines(data)
	previous := ""
	hasPrevious := false
	for index := 0; index < len(lines); index++ {
		if !hasPrevious || lines[index] != previous {
			fmt.Print(lines[index], "\n")
			previous = lines[index]
			hasPrevious = true
		}
	}
	if status != 0 {
		os.Exit(status)
	}
}
