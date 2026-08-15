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

func reverse(line string) {
	starts := []int{}
	for index := 0; index < len(line); {
		starts = append(starts, index)
		value := line[index]
		width := 1
		if value >= 240 {
			width = 4
		} else if value >= 224 {
			width = 3
		} else if value >= 192 {
			width = 2
		}
		index = index + width
	}
	for index := len(starts); index > 0; index-- {
		start := starts[index-1]
		limit := len(line)
		if index < len(starts) {
			limit = starts[index]
		}
		fmt.Print(line[start:limit])
	}
	fmt.Print("\n")
}

func main() {
	data, status := userland.ReadInput("rev", os.Args[1:])
	lines := splitLines(data)
	for index := 0; index < len(lines); index++ {
		reverse(lines[index])
	}
	if status != 0 {
		os.Exit(status)
	}
}
