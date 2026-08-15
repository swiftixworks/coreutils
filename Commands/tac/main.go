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
	data, status := userland.ReadInput("tac", os.Args[1:])
	lines := splitLines(data)
	for index := len(lines); index > 0; index-- {
		fmt.Print(lines[index-1])
		if index > 1 || len(data) == 0 || data[len(data)-1] == 10 {
			fmt.Print("\n")
		}
	}
	if status != 0 {
		os.Exit(status)
	}
}
