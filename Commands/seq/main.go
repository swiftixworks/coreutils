package main

import "fmt"
import "os"
import "strconv"

func usage() {
	fmt.Println("seq: usage: seq [start [step]] stop")
	os.Exit(2)
}

func number(index int) int {
	value, err := strconv.Atoi(os.Args[index])
	if err != nil {
		usage()
	}
	return value
}

func main() {
	count := len(os.Args) - 1
	if count < 1 || count > 3 {
		usage()
	}

	start := 1
	step := 1
	stop := 0
	if count == 1 {
		stop = number(1)
	} else if count == 2 {
		start = number(1)
		stop = number(2)
	} else {
		start = number(1)
		step = number(2)
		stop = number(3)
	}

	if step == 0 {
		fmt.Println("seq: step cannot be 0")
		os.Exit(2)
	}
	value := start
	for (step > 0 && value <= stop) || (step < 0 && value >= stop) {
		fmt.Println(value)
		value = value + step
	}
}
