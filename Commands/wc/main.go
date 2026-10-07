package main

import "fmt"
import "os"
import "strconv"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: wc [OPTION]... [FILE]...\nPrint newline, word, and byte counts for each FILE, and a total line if\nmore than one FILE is specified. A word is a non-empty sequence of\nnon-white-space characters.\n\nWith no FILE, or when FILE is -, read standard input.\n\n  -c, --bytes              print the byte counts\n  -m, --chars              print the character counts\n  -l, --lines              print the newline counts\n  -L, --max-line-length    print the maximum display width\n  -w, --words              print the word counts\n      --help               display this help and exit\n"

// wordCount counts white-space separated words with whole-buffer passes.
func wordCount(data string) int {
	text := cli.ReplaceAll(data, "\n", " ")
	text = cli.ReplaceAll(text, "\t", " ")
	text = cli.ReplaceAll(text, "\r", " ")
	text = cli.ReplaceAll(text, "\x0b", " ")
	text = cli.ReplaceAll(text, "\x0c", " ")
	text = strings.TrimSpace(cli.SqueezeRuns(text, " "))
	if text == "" {
		return 0
	}
	return strings.Count(text, " ") + 1
}

// lineWidth is the display width of one line: characters, with tabs advancing
// to the next multiple of eight.
func lineWidth(line string, ascii bool) int {
	if !strings.Contains(line, "\t") {
		if ascii {
			return len(line)
		}
		return cli.CharCount(line)
	}
	pieces := strings.Split(line, "\t")
	width := 0
	for index := 0; index < len(pieces); index++ {
		if index > 0 {
			width = (width/8 + 1) * 8
		}
		width = width + cli.CharCount(pieces[index])
	}
	return width
}

func maxLineWidth(data string) int {
	lines := cli.SplitLines(data)
	widest := 0
	if cli.IsASCII(data) && !strings.Contains(data, "\t") {
		for index := 0; index < len(lines); index++ {
			if len(lines[index]) > widest {
				widest = len(lines[index])
			}
		}
		return widest
	}
	ascii := cli.IsASCII(data)
	for index := 0; index < len(lines); index++ {
		width := lineWidth(lines[index], ascii)
		if width > widest {
			widest = width
		}
	}
	return widest
}

func main() {
	flags, values, operands := cli.ParseOptions("wc", usage, os.Args[1:], "lwcmL", "lines=l words=w bytes=c chars=m max-line-length=L", cli.ModePlain, 1)
	showLines := false
	showWords := false
	showBytes := false
	showChars := false
	showWidth := false
	for index := 0; index < len(flags); index++ {
		flag := flags[index] + values[index]
		if flag == "l" {
			showLines = true
		} else if flag == "w" {
			showWords = true
		} else if flag == "c" {
			showBytes = true
		} else if flag == "m" {
			showChars = true
		} else {
			showWidth = true
		}
	}
	if len(flags) == 0 {
		showLines = true
		showWords = true
		showBytes = true
	}
	columns := 0
	if showLines {
		columns++
	}
	if showWords {
		columns++
	}
	if showBytes {
		columns++
	}
	if showChars {
		columns++
	}
	if showWidth {
		columns++
	}

	names := []string{}
	counts := []int{}
	totals := []int{0, 0, 0, 0, 0}
	implicit := len(operands) == 0
	stdin := implicit
	if implicit {
		operands = []string{"-"}
	}
	result := 0
	for index := 0; index < len(operands); index++ {
		data, status := cli.ReadFile("wc", operands[index])
		if status != 0 {
			result = 1
			continue
		}
		row := []int{0, 0, 0, 0, 0}
		if showLines {
			row[0] = strings.Count(data, "\n")
		}
		if showWords {
			row[1] = wordCount(data)
		}
		if showChars {
			row[2] = cli.CharCount(data)
		}
		if showBytes {
			row[3] = len(data)
		}
		if showWidth {
			row[4] = maxLineWidth(data)
		}
		if operands[index] == "-" {
			stdin = true
		}
		if implicit {
			names = append(names, "")
		} else {
			names = append(names, operands[index])
		}
		for column := 0; column < 5; column++ {
			counts = append(counts, row[column])
			if column == 4 {
				if row[4] > totals[4] {
					totals[4] = row[4]
				}
			} else {
				totals[column] = totals[column] + row[column]
			}
		}
	}
	if len(operands) > 1 {
		names = append(names, "total")
		for column := 0; column < 5; column++ {
			counts = append(counts, totals[column])
		}
	}

	width := 1
	if columns > 1 || len(operands) > 1 {
		for index := 0; index < len(counts); index++ {
			digits := cli.DigitCount(counts[index])
			if digits > width {
				width = digits
			}
		}
		if stdin && width < 7 {
			width = 7
		}
	}
	show := []bool{showLines, showWords, showChars, showBytes, showWidth}
	report := ""
	for row := 0; row < len(names); row++ {
		first := true
		for column := 0; column < 5; column++ {
			if !show[column] {
				continue
			}
			if !first {
				report = report + " "
			}
			report = report + cli.PadLeft(strconv.Itoa(counts[row*5+column]), width)
			first = false
		}
		if names[row] != "" {
			report = report + " " + names[row]
		}
		report = report + "\n"
	}
	fmt.Print(report)
	os.Exit(result)
}
