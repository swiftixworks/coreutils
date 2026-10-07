package main

import "fmt"
import "os"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: rev [FILE]...\nReverse the characters of every line of each FILE.\n\nWith no FILE, or when FILE is -, read standard input.\n\n      --help   display this help and exit\n"

func reverse(line string) string {
	characters := strings.Split(line, "")
	count := len(characters)
	for index := 0; index < count/2; index++ {
		other := count - 1 - index
		character := characters[index]
		characters[index] = characters[other]
		characters[other] = character
	}
	return strings.Join(characters, "")
}

func main() {
	flags, values, operands := cli.ParseOptions("rev", usage, os.Args[1:], "", "", cli.ModePlain, 1)
	if len(flags) != len(values) {
		os.Exit(1)
	}
	data, status := cli.ReadAll("rev", operands)
	terminated := strings.HasSuffix(data, "\n")
	lines := cli.SplitLines(data)
	for index := 0; index < len(lines); index++ {
		if len(lines[index]) > 1 {
			lines[index] = reverse(lines[index])
		}
	}
	text := strings.Join(lines, "\n")
	if terminated {
		text = text + "\n"
	}
	fmt.Print(text)
	os.Exit(status)
}
