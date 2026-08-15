package main

import "fmt"
import "os"
import "strconv"
import "swiftix/userland"

func usage() {
	fmt.Println("tail: usage: tail [-n count] [file...]")
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
	data, status := userland.ReadInput("tail", os.Args[first:])
	starts := []int{0}
	for index := 0; index < len(data); index++ {
		if data[index] == 10 && index+1 < len(data) {
			starts = append(starts, index+1)
		}
	}
	start := 0
	if count == 0 {
		start = len(data)
	} else if len(starts) > count {
		start = starts[len(starts)-count]
	}
	fmt.Print(data[start:])
	if status != 0 {
		os.Exit(status)
	}
}
