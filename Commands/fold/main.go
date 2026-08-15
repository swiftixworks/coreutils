package main

import "fmt"
import "os"
import "strconv"
import "swiftix/userland"

func usage() {
	fmt.Println("fold: usage: fold [-w width] [file...]")
	os.Exit(2)
}

func runeWidth(value int) int {
	if value >= 240 {
		return 4
	}
	if value >= 224 {
		return 3
	}
	if value >= 192 {
		return 2
	}
	return 1
}

func main() {
	width := 80
	first := 1
	if len(os.Args) >= 3 && (os.Args[1] == "-w" || os.Args[1] == "--width") {
		value, err := strconv.Atoi(os.Args[2])
		if err != nil || value <= 0 {
			usage()
		}
		width = value
		first = 3
	}

	data, status := userland.ReadInput("fold", os.Args[first:])
	column := 0
	for index := 0; index < len(data); {
		if data[index] == 10 {
			fmt.Print("\n")
			column = 0
			index++
			continue
		}
		if column == width {
			fmt.Print("\n")
			column = 0
		}
		count := runeWidth(data[index])
		if index+count > len(data) {
			count = 1
		}
		fmt.Print(data[index:index+count])
		index = index + count
		column++
	}
	if status != 0 {
		os.Exit(status)
	}
}
