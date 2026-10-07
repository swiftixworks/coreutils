package main

import "fmt"
import "os"
import "strconv"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: nl [OPTION]... [FILE]...\nWrite each FILE to standard output, with line numbers added.\n\nWith no FILE, or when FILE is -, read standard input.\n\n  -b, --body-numbering=STYLE      use STYLE for numbering body lines:\n                                    a (all lines), t (nonempty lines, default),\n                                    n (no lines)\n  -i, --line-increment=NUMBER     line number increment at each line (default 1)\n  -n, --number-format=FORMAT      insert line numbers according to FORMAT:\n                                    ln (left justified), rn (right justified,\n                                    default), rz (right justified, leading zeros)\n  -s, --number-separator=STRING   add STRING after the line number (default TAB)\n  -v, --starting-line-number=NUMBER  first line number (default 1)\n  -w, --number-width=NUMBER       use NUMBER columns for line numbers (default 6)\n      --help                      display this help and exit\n"

func integer(text string, what string) int {
	value, err := strconv.Atoi(text)
	if err != nil {
		cli.Fail("nl", "invalid "+what+": '"+text+"'", 1)
	}
	return value
}

// numbered formats one line number: ln left-justified, rn right-justified,
// rz right-justified with leading zeros after the sign.
func numbered(number int, width int, format string) string {
	text := strconv.Itoa(number)
	missing := width - len(text)
	if missing <= 0 {
		return text
	}
	if format == "ln" {
		return text + strings.Repeat(" ", missing)
	}
	if format == "rz" {
		if number < 0 {
			return "-" + strings.Repeat("0", missing) + text[1:]
		}
		return strings.Repeat("0", missing) + text
	}
	return strings.Repeat(" ", missing) + text
}

func main() {
	flags, values, operands := cli.ParseOptions("nl", usage, os.Args[1:], "b:i:n:s:v:w:", "body-numbering=b line-increment=i number-format=n number-separator=s starting-line-number=v number-width=w", cli.ModePlain, 1)
	style := "t"
	format := "rn"
	separator := "\t"
	start := 1
	step := 1
	width := 6
	for index := 0; index < len(flags); index++ {
		flag := flags[index]
		value := values[index]
		if flag == "b" {
			if value != "a" && value != "t" && value != "n" {
				cli.Fail("nl", "invalid body numbering style: '"+value+"'", 1)
			}
			style = value
		} else if flag == "i" {
			step = integer(value, "line number increment")
		} else if flag == "n" {
			if value != "ln" && value != "rn" && value != "rz" {
				cli.Fail("nl", "invalid line numbering format: '"+value+"'", 1)
			}
			format = value
		} else if flag == "s" {
			separator = value
		} else if flag == "v" {
			start = integer(value, "starting line number")
		} else {
			width = integer(value, "line number field width")
			if width < 1 {
				cli.Fail("nl", "invalid line number field width: '"+value+"'", 1)
			}
		}
	}

	data, status := cli.ReadAll("nl", operands)
	terminated := strings.HasSuffix(data, "\n")
	lines := cli.SplitLines(data)
	blank := strings.Repeat(" ", width+len(separator))
	// Right-aligned non-negative numbers, the usual case, are padded in the
	// loop; a helper call per line is far more expensive than these few steps.
	fill := strings.Repeat(" ", width)
	if format == "rz" {
		fill = strings.Repeat("0", width)
	}
	rightAligned := format != "ln"
	none := style == "n"
	nonempty := style == "t"
	number := start
	for index := 0; index < len(lines); index++ {
		if none || (nonempty && lines[index] == "") {
			lines[index] = blank + lines[index]
			continue
		}
		text := strconv.Itoa(number)
		if len(text) < width {
			if rightAligned && number >= 0 {
				text = fill[len(text):] + text
			} else {
				text = numbered(number, width, format)
			}
		}
		lines[index] = text + separator + lines[index]
		number = number + step
	}
	text := strings.Join(lines, "\n")
	if terminated {
		text = text + "\n"
	}
	fmt.Print(text)
	os.Exit(status)
}
