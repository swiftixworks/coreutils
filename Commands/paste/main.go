package main

import "fmt"
import "os"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: paste [OPTION]... [FILE]...\nWrite lines consisting of the sequentially corresponding lines from each\nFILE, separated by TABs, to standard output.\n\nWith no FILE, or when FILE is -, read standard input.\n\n  -d, --delimiters=LIST   reuse characters from LIST instead of TABs\n  -s, --serial            paste one file at a time instead of in parallel\n      --help              display this help and exit\n\nLIST may contain the escapes \\n, \\t, \\\\ and \\0 (empty string).\n"

// delimiterList expands the -d LIST into one string per delimiter.
func delimiterList(list string) []string {
	characters := strings.Split(list, "")
	result := []string{}
	for index := 0; index < len(characters); index++ {
		character := characters[index]
		if character == "\\" && index+1 < len(characters) {
			index++
			character = characters[index]
			if character == "n" {
				character = "\n"
			} else if character == "t" {
				character = "\t"
			} else if character == "0" {
				character = ""
			}
		}
		result = append(result, character)
	}
	return result
}

func main() {
	flags, values, operands := cli.ParseOptions("paste", usage, os.Args[1:], "d:s", "delimiters=d serial=s", cli.ModePlain, 1)
	delimiters := []string{"\t"}
	serial := false
	for index := 0; index < len(flags); index++ {
		if flags[index] == "s" {
			serial = true
		} else {
			delimiters = delimiterList(values[index])
			if len(delimiters) == 0 {
				delimiters = []string{""}
			}
		}
	}
	if len(operands) == 0 {
		operands = []string{"-"}
	}

	// Every "-" operand draws from the same standard-input line stream.
	stdinLines := []string{}
	stdinLoaded := false
	stdinNextLine := 0
	result := 0
	columns := len(operands)
	sources := make([][]string, columns)
	shared := make([]bool, columns)
	for index := 0; index < columns; index++ {
		if operands[index] == "-" {
			if !stdinLoaded {
				stdinLines = cli.SplitLines(cli.ReadStdin())
				stdinLoaded = true
			}
			shared[index] = true
			continue
		}
		data, status := cli.ReadFile("paste", operands[index])
		if status != 0 {
			result = 1
		}
		sources[index] = cli.SplitLines(data)
	}

	if serial {
		for index := 0; index < columns; index++ {
			lines := sources[index]
			if shared[index] {
				lines = stdinLines[stdinNextLine:]
				stdinNextLine = len(stdinLines)
			}
			if len(delimiters) == 1 {
				fmt.Print(strings.Join(lines, delimiters[0]) + "\n")
				continue
			}
			for position := 1; position < len(lines); position++ {
				lines[position] = delimiters[(position-1)%len(delimiters)] + lines[position]
			}
			fmt.Print(strings.Join(lines, "") + "\n")
		}
		os.Exit(result)
	}

	positions := make([]int, columns)
	rows := []string{}
	for {
		row := ""
		more := false
		for index := 0; index < columns; index++ {
			if index > 0 {
				row = row + delimiters[(index-1)%len(delimiters)]
			}
			if shared[index] {
				if stdinNextLine < len(stdinLines) {
					row = row + stdinLines[stdinNextLine]
					stdinNextLine++
					more = true
				}
			} else if positions[index] < len(sources[index]) {
				row = row + sources[index][positions[index]]
				positions[index] = positions[index] + 1
				more = true
			}
		}
		if !more {
			break
		}
		rows = append(rows, row)
	}
	if len(rows) > 0 {
		fmt.Print(strings.Join(rows, "\n") + "\n")
	}
	os.Exit(result)
}
