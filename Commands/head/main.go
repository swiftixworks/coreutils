package main

import "fmt"
import "os"
import "strconv"
import "swiftix/userland"

func usage() {
	fmt.Println("head: usage: head [-n count] [file...]")
	os.Exit(2)
}

func options() (int, int) {
	count := 10
	first := 1
	if len(os.Args) >= 3 && os.Args[1] == "-n" {
		value, err := strconv.Atoi(os.Args[2])
		if err != nil || value < 0 {
			usage()
		}
		count = value
		first = 3
	} else if len(os.Args) >= 2 && len(os.Args[1]) > 1 && os.Args[1][0] == 45 {
		value, err := strconv.Atoi(os.Args[1][1:])
		if err != nil || value < 0 {
			usage()
		}
		count = value
		first = 2
	}
	return count, first
}

func main() {
	count, first := options()
	data, status := userland.ReadInput("head", os.Args[first:])
	limit := 0
	lines := 0
	for limit < len(data) {
		if lines >= count {
			break
		}
		if data[limit] == 10 {
			lines++
		}
		limit++
	}
	fmt.Print(data[:limit])
	if status != 0 {
		os.Exit(status)
	}
}
