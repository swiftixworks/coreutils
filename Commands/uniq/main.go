package main

import "fmt"
import "os"
import "strconv"
import "strings"
import "swiftix/userland"
import "swiftix/internal/cli"

const usage = "Usage: uniq [OPTION]... [INPUT [OUTPUT]]\nFilter adjacent matching lines from INPUT (or standard input),\nwriting to OUTPUT (or standard output).\n\nWith no options, matching lines are merged to the first occurrence.\n\n  -c, --count           prefix lines by the number of occurrences\n  -d, --repeated        only print duplicate lines, one for each group\n  -f, --skip-fields=N   avoid comparing the first N fields\n  -i, --ignore-case     ignore differences in case when comparing\n  -s, --skip-chars=N    avoid comparing the first N characters\n  -u, --unique          only print unique lines\n  -w, --check-chars=N   compare no more than N characters in lines\n      --help            display this help and exit\n\nA field is a run of blanks (usually spaces and/or TABs), then non-blank\ncharacters. Fields are skipped before chars.\n"

// compareKey is the part of line that takes part in comparison.
func compareKey(line string, fields int, skip int, check int) string {
	position := 0
	for field := 0; field < fields; field++ {
		for position < len(line) && (line[position] == 32 || line[position] == 9) {
			position++
		}
		for position < len(line) && line[position] != 32 && line[position] != 9 {
			position++
		}
	}
	rest := line[position:]
	if skip == 0 && check < 0 {
		return rest
	}
	if cli.IsASCII(rest) {
		if skip > len(rest) {
			skip = len(rest)
		}
		rest = rest[skip:]
		if check >= 0 && check < len(rest) {
			rest = rest[:check]
		}
		return rest
	}
	characters := strings.Split(rest, "")
	if skip > len(characters) {
		skip = len(characters)
	}
	characters = characters[skip:]
	if check >= 0 && check < len(characters) {
		characters = characters[:check]
	}
	return strings.Join(characters, "")
}

func number(text string, what string) int {
	value, ok := cli.ParseCount(text)
	if !ok {
		cli.Fail("uniq", text+": invalid number of "+what, 1)
	}
	return value
}

func main() {
	flags, values, operands := cli.ParseOptions("uniq", usage, os.Args[1:], "cdf:is:uw:", "count=c repeated=d skip-fields=f ignore-case=i skip-chars=s unique=u check-chars=w", cli.ModePlain, 1)
	count := false
	repeated := false
	unique := false
	ignoreCase := false
	fields := 0
	skip := 0
	check := 0 - 1
	for index := 0; index < len(flags); index++ {
		flag := flags[index]
		if flag == "c" {
			count = true
		} else if flag == "d" {
			repeated = true
		} else if flag == "u" {
			unique = true
		} else if flag == "i" {
			ignoreCase = true
		} else if flag == "f" {
			fields = number(values[index], "fields to skip")
		} else if flag == "s" {
			skip = number(values[index], "bytes to skip")
		} else {
			check = number(values[index], "bytes to compare")
		}
	}
	if len(operands) > 2 {
		cli.FailUsage("uniq", "extra operand '"+operands[2]+"'", 1)
	}
	input := "-"
	if len(operands) >= 1 {
		input = operands[0]
	}
	outputPath := ""
	if len(operands) == 2 && operands[1] != "-" {
		outputPath = operands[1]
	}
	data, status := cli.ReadFile("uniq", input)
	if status != 0 {
		os.Exit(1)
	}

	lines := cli.SplitLines(data)
	keys := lines
	if ignoreCase {
		keys = cli.SplitLines(cli.UpperASCII(data))
	}
	if fields > 0 || skip > 0 || check >= 0 {
		compared := make([]string, len(keys))
		for index := 0; index < len(keys); index++ {
			compared[index] = compareKey(keys[index], fields, skip, check)
		}
		keys = compared
	}

	total := len(lines)
	single := !repeated
	several := !unique
	blanks := "       "
	kept := []string{}
	start := 0
	for index := 1; index <= total; index++ {
		if index < total && keys[index] == keys[start] {
			continue
		}
		occurrences := index - start
		if (occurrences == 1 && single) || (occurrences > 1 && several) {
			if count {
				number := strconv.Itoa(occurrences)
				if len(number) < 7 {
					number = blanks[len(number):] + number
				}
				kept = append(kept, number+" "+lines[start])
			} else {
				kept = append(kept, lines[start])
			}
		}
		start = index
	}
	text := ""
	if len(kept) > 0 {
		text = strings.Join(kept, "\n") + "\n"
	}
	if outputPath == "" {
		fmt.Print(text)
	} else if userland.WriteFile(outputPath, text) != 0 {
		cli.Fail("uniq", outputPath+": cannot write", 1)
	}
	os.Exit(0)
}
