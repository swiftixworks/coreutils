package main

import "fmt"
import "os"
import "swiftix/userland"

func main() {
	data, status := userland.ReadInput("wc", os.Args[1:])
	lines := 0
	words := 0
	inWord := false
	for index := 0; index < len(data); index++ {
		value := data[index]
		if value == 10 {
			lines++
		}
		space := value == 32 || value == 9 || value == 10 || value == 13
		if space {
			inWord = false
		} else if !inWord {
			words++
			inWord = true
		}
	}
	fmt.Println(lines, words, len(data))
	if status != 0 {
		os.Exit(status)
	}
}
