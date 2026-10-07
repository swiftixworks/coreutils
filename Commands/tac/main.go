package main

import "fmt"
import "os"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: tac [OPTION]... [FILE]...\nWrite each FILE to standard output, last line first.\n\nWith no FILE, or when FILE is -, read standard input.\n\n  -s, --separator=STRING   use STRING as the separator instead of newline\n      --help               display this help and exit\n"

func main() {
	flags, values, operands := cli.ParseOptions("tac", usage, os.Args[1:], "s:", "separator=s", cli.ModePlain, 1)
	separator := "\n"
	for index := 0; index < len(flags); index++ {
		if flags[index] == "s" {
			separator = values[index]
		}
	}
	if separator == "" {
		cli.Fail("tac", "separator cannot be empty", 1)
	}
	if len(operands) == 0 {
		operands = []string{"-"}
	}
	result := 0
	for index := 0; index < len(operands); index++ {
		data, status := cli.ReadFile("tac", operands[index])
		if status != 0 {
			result = 1
			continue
		}
		// Records keep their trailing separator; a final record without one is
		// written first, unterminated, exactly as GNU tac does.
		tail := ""
		if !strings.HasSuffix(data, separator) {
			cut := strings.LastIndex(data, separator)
			if cut < 0 {
				fmt.Print(data)
				continue
			}
			tail = data[cut+len(separator):]
			data = data[:cut+len(separator)]
		}
		records := strings.Split(data[:len(data)-len(separator)], separator)
		count := len(records)
		reversed := make([]string, count)
		for position := 0; position < count; position++ {
			reversed[position] = records[count-1-position]
		}
		fmt.Print(tail + strings.Join(reversed, separator) + separator)
	}
	os.Exit(result)
}
