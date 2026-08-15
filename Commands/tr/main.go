package main

import "fmt"
import "os"
import "swiftix/userland"

func usage() {
	fmt.Println("tr: usage: tr [-ds] set1 [set2]")
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

func asciiCharacter(value int) string {
	digits := "0123456789"
	upper := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower := "abcdefghijklmnopqrstuvwxyz"
	if value >= 48 && value <= 57 {
		return digits[value-48:value-47]
	}
	if value >= 65 && value <= 90 {
		return upper[value-65:value-64]
	}
	if value >= 97 && value <= 122 {
		return lower[value-97:value-96]
	}
	return ""
}

func characters(value string) []string {
	result := []string{}
	for index := 0; index < len(value); {
		if index+2 < len(value) && value[index+1] == 45 && value[index] < 128 && value[index+2] < 128 && value[index] <= value[index+2] {
			start := value[index]
			stop := value[index+2]
			first := asciiCharacter(start)
			last := asciiCharacter(stop)
			if first != "" && last != "" {
				for current := start; current <= stop; current++ {
					character := asciiCharacter(current)
					if character != "" {
						result = append(result, character)
					}
				}
				index = index + 3
				continue
			}
		}
		width := runeWidth(value[index])
		if index+width > len(value) {
			width = 1
		}
		result = append(result, value[index:index+width])
		index = index + width
	}
	return result
}

func position(values []string, target string) int {
	for index := 0; index < len(values); index++ {
		if values[index] == target {
			return index
		}
	}
	return -1
}

func main() {
	deleteValues := false
	squeeze := false
	first := 1
	if len(os.Args) > 1 && len(os.Args[1]) > 1 && os.Args[1][0] == 45 {
		for index := 1; index < len(os.Args[1]); index++ {
			if os.Args[1][index] == 100 {
				deleteValues = true
			} else if os.Args[1][index] == 115 {
				squeeze = true
			} else {
				usage()
			}
		}
		first = 2
	}
	required := 2
	if deleteValues || squeeze {
		required = 1
	}
	if len(os.Args)-first < required || len(os.Args)-first > 2 {
		usage()
	}

	from := characters(os.Args[first])
	to := []string{}
	if first+1 < len(os.Args) {
		to = characters(os.Args[first+1])
	}
	data, status := userland.ReadInput("tr", []string{})
	last := ""
	hasLast := false
	for index := 0; index < len(data); {
		width := runeWidth(data[index])
		if index+width > len(data) {
			width = 1
		}
		value := data[index:index+width]
		index = index + width
		fromIndex := position(from, value)
		if deleteValues && fromIndex >= 0 {
			continue
		}
		if !deleteValues && fromIndex >= 0 && len(to) > 0 {
			toIndex := fromIndex
			if toIndex >= len(to) {
				toIndex = len(to) - 1
			}
			value = to[toIndex]
		}
		squeezeSet := from
		if len(to) > 0 {
			squeezeSet = to
		}
		if squeeze && hasLast && value == last && position(squeezeSet, value) >= 0 {
			continue
		}
		fmt.Print(value)
		last = value
		hasLast = true
	}
	if status != 0 {
		os.Exit(status)
	}
}
