package main

import "fmt"
import "os"
import "strconv"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: comm [OPTION]... FILE1 FILE2\nCompare sorted files FILE1 and FILE2 line by line.\n\nWhen FILE1 or FILE2 (not both) is -, read standard input.\n\nWith no options, produce three-column output. Column one contains\nlines unique to FILE1, column two contains lines unique to FILE2,\nand column three contains lines common to both files.\n\n  -1                         suppress column 1 (lines unique to FILE1)\n  -2                         suppress column 2 (lines unique to FILE2)\n  -3                         suppress column 3 (lines that appear in both files)\n      --output-delimiter=STR separate columns with STR\n      --total                output a summary\n      --help                 display this help and exit\n"

func main() {
	flags, values, operands := cli.ParseOptions("comm", usage, os.Args[1:], "123d:t", "output-delimiter=d total=t", cli.ModePlain, 1)
	showFirst := true
	showSecond := true
	showBoth := true
	delimiter := "\t"
	total := false
	for index := 0; index < len(flags); index++ {
		flag := flags[index]
		if flag == "1" {
			showFirst = false
		} else if flag == "2" {
			showSecond = false
		} else if flag == "3" {
			showBoth = false
		} else if flag == "d" {
			delimiter = values[index]
		} else {
			total = true
		}
	}
	if len(operands) < 2 {
		if len(operands) == 0 {
			cli.FailUsage("comm", "missing operand", 1)
		}
		cli.FailUsage("comm", "missing operand after '"+operands[0]+"'", 1)
	}
	if len(operands) > 2 {
		cli.FailUsage("comm", "extra operand '"+operands[2]+"'", 1)
	}
	firstData, firstStatus := cli.ReadFile("comm", operands[0])
	secondData, secondStatus := cli.ReadFile("comm", operands[1])
	if firstStatus != 0 || secondStatus != 0 {
		os.Exit(1)
	}
	first := cli.SplitLines(firstData)
	second := cli.SplitLines(secondData)

	secondPrefix := ""
	if showFirst {
		secondPrefix = delimiter
	}
	bothPrefix := secondPrefix
	if showSecond {
		bothPrefix = bothPrefix + delimiter
	}
	left := 0
	right := 0
	onlyFirst := 0
	onlySecond := 0
	both := 0
	rows := []string{}
	for left < len(first) || right < len(second) {
		if right >= len(second) {
			onlyFirst++
			if showFirst {
				rows = append(rows, first[left])
			}
			left++
		} else if left >= len(first) {
			onlySecond++
			if showSecond {
				rows = append(rows, secondPrefix+second[right])
			}
			right++
		} else if first[left] == second[right] {
			both++
			if showBoth {
				rows = append(rows, bothPrefix+first[left])
			}
			left++
			right++
		} else if first[left] < second[right] {
			onlyFirst++
			if showFirst {
				rows = append(rows, first[left])
			}
			left++
		} else {
			onlySecond++
			if showSecond {
				rows = append(rows, secondPrefix+second[right])
			}
			right++
		}
	}
	if total {
		rows = append(rows, strconv.Itoa(onlyFirst)+delimiter+strconv.Itoa(onlySecond)+delimiter+strconv.Itoa(both)+delimiter+"total")
	}
	if len(rows) > 0 {
		fmt.Print(strings.Join(rows, "\n") + "\n")
	}
	os.Exit(0)
}
