package main

import "fmt"
import "os"
import "swiftix/userland"

func usage() {
	fmt.Println("paste: usage: paste [-d delimiter] [file...]")
	os.Exit(2)
}

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
	delimiter := "\t"
	first := 1
	if len(os.Args) >= 2 && os.Args[1] == "-d" {
		if len(os.Args) < 3 || len(os.Args[2]) == 0 {
			usage()
		}
		delimiter = os.Args[2]
		first = 3
	}

	if first == len(os.Args) {
		data, status := userland.ReadInput("paste", []string{})
		fmt.Print(data)
		if status != 0 {
			os.Exit(status)
		}
		return
	}

	inputs := [][]string{}
	status := 0
	maxLines := 0
	for index := first; index < len(os.Args); index++ {
		data, readStatus := userland.ReadInput("paste", []string{os.Args[index]})
		lines := splitLines(data)
		inputs = append(inputs, lines)
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
		if readStatus != 0 {
			status = readStatus
		}
	}

	for line := 0; line < maxLines; line++ {
		for input := 0; input < len(inputs); input++ {
			if input > 0 {
				fmt.Print(delimiter)
			}
			if line < len(inputs[input]) {
				fmt.Print(inputs[input][line])
			}
		}
		fmt.Print("\n")
	}
	if status != 0 {
		os.Exit(status)
	}
}
