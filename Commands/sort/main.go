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
	first := 1
	reverse := false
	if len(os.Args) > 1 && os.Args[1] == "-r" {
		reverse = true
		first = 2
	}
	data, status := userland.ReadInput("sort", os.Args[first:])
	lines := splitLines(data)
	for index := 1; index < len(lines); index++ {
		value := lines[index]
		position := index
		for position > 0 {
			if lines[position-1] <= value {
				break
			}
			lines[position] = lines[position-1]
			position--
		}
		lines[position] = value
	}
	if reverse {
		for index := len(lines); index > 0; index-- {
			fmt.Print(lines[index-1], "\n")
		}
	} else {
		for index := 0; index < len(lines); index++ {
			fmt.Print(lines[index], "\n")
		}
	}
	if status != 0 {
		os.Exit(status)
	}
}
