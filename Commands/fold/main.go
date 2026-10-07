package main

import "fmt"
import "os"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: fold [OPTION]... [FILE]...\nWrap input lines in each FILE, writing to standard output.\n\nWith no FILE, or when FILE is -, read standard input.\n\n  -b, --bytes         count bytes rather than characters\n  -s, --spaces        break at spaces\n  -w, --width=WIDTH   use WIDTH columns instead of 80\n  -WIDTH              same as -w WIDTH\n      --help          display this help and exit\n"

// charBoundary moves offset back to the nearest character boundary of line so
// a row never ends inside a UTF-8 sequence.
func charBoundary(line string, offset int) int {
	for offset > 0 && offset < len(line) && line[offset] >= 128 && line[offset] < 192 {
		offset--
	}
	return offset
}

// foldLine wraps one over-long line. With spaces set, a row ends after its last
// blank when it has one.
func foldLine(line string, width int, spaces bool, bytes bool) string {
	if bytes || cli.IsASCII(line) {
		rows := ""
		for len(line) > width {
			cut := width
			if spaces {
				blank := strings.LastIndex(line[:width], " ")
				if blank >= 0 {
					cut = blank + 1
				}
			}
			if bytes {
				cut = charBoundary(line, cut)
				if cut == 0 {
					cut = width
				}
			}
			rows = rows + line[:cut] + "\n"
			line = line[cut:]
		}
		return rows + line
	}
	characters := strings.Split(line, "")
	rows := ""
	start := 0
	for len(characters)-start > width {
		cut := width
		if spaces {
			for position := width - 1; position >= 0; position-- {
				if characters[start+position] == " " {
					cut = position + 1
					break
				}
			}
		}
		rows = rows + strings.Join(characters[start:start+cut], "") + "\n"
		start = start + cut
	}
	return rows + strings.Join(characters[start:], "")
}

func main() {
	flags, values, operands := cli.ParseOptions("fold", usage, os.Args[1:], "bsw:", "bytes=b spaces=s width=w", cli.ModeCount, 1)
	width := 80
	spaces := false
	bytes := false
	for index := 0; index < len(flags); index++ {
		flag := flags[index]
		if flag == "b" {
			bytes = true
		} else if flag == "s" {
			spaces = true
		} else {
			parsed, ok := cli.ParseCount(values[index])
			if !ok || parsed == 0 {
				cli.Fail("fold", "invalid number of columns: '"+values[index]+"'", 1)
			}
			width = parsed
		}
	}

	data, status := cli.ReadAll("fold", operands)
	terminated := strings.HasSuffix(data, "\n")
	lines := cli.SplitLines(data)
	ascii := bytes || cli.IsASCII(data)
	// Only an over-long line is rebuilt; the others pass through as they are.
	for index := 0; index < len(lines); index++ {
		line := lines[index]
		if len(line) <= width {
			continue
		}
		if !ascii && cli.CharCount(line) <= width {
			continue
		}
		lines[index] = foldLine(line, width, spaces, bytes)
	}
	text := strings.Join(lines, "\n")
	if terminated {
		text = text + "\n"
	}
	fmt.Print(text)
	os.Exit(status)
}
