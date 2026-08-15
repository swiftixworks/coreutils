package main

import "fmt"
import "os"
import "swiftix/userland"

func main() {
	data, status := userland.ReadInput("cat", os.Args[1:])
	fmt.Print(data)
	if status != 0 {
		os.Exit(status)
	}
}
