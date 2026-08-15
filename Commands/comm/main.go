package main

import "fmt"
import "os"
import "swiftix/userland"

func usage() {
	fmt.Println("comm: usage: comm [-123] file1 file2")
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

func hasOption(options string, value int) bool {
	for index := 1; index < len(options); index++ {
		if options[index] == value {
			return true
		}
		if options[index] != 49 && options[index] != 50 && options[index] != 51 {
			usage()
		}
	}
	return false
}

func main() {
	first := 1
	hide1 := false
	hide2 := false
	hide3 := false
	for first < len(os.Args) && len(os.Args[first]) > 1 && os.Args[first][0] == 45 {
		if hasOption(os.Args[first], 49) {
			hide1 = true
		}
		if hasOption(os.Args[first], 50) {
			hide2 = true
		}
		if hasOption(os.Args[first], 51) {
			hide3 = true
		}
		first++
	}
	if len(os.Args)-first != 2 {
		usage()
	}

	leftData, leftStatus := userland.ReadInput("comm", []string{os.Args[first]})
	rightData, rightStatus := userland.ReadInput("comm", []string{os.Args[first+1]})
	if leftStatus != 0 || rightStatus != 0 {
		os.Exit(1)
	}
	left := splitLines(leftData)
	right := splitLines(rightData)
	leftIndex := 0
	rightIndex := 0
	for leftIndex < len(left) || rightIndex < len(right) {
		if rightIndex >= len(right) || (leftIndex < len(left) && left[leftIndex] < right[rightIndex]) {
			if !hide1 {
				fmt.Println(left[leftIndex])
			}
			leftIndex++
		} else if leftIndex >= len(left) || right[rightIndex] < left[leftIndex] {
			if !hide2 {
				if !hide1 {
					fmt.Print("\t")
				}
				fmt.Println(right[rightIndex])
			}
			rightIndex++
		} else {
			if !hide3 {
				if !hide1 {
					fmt.Print("\t")
				}
				if !hide2 {
					fmt.Print("\t")
				}
				fmt.Println(left[leftIndex])
			}
			leftIndex++
			rightIndex++
		}
	}
}
