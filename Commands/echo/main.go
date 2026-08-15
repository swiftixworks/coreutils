package main

import "fmt"
import "os"

func main() {
	for index := 1; index < len(os.Args); index++ {
		if index > 1 {
			fmt.Print(" ")
		}
		fmt.Print(os.Args[index])
	}
	fmt.Print("\n")
}
