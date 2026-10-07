package main

import "fmt"
import "os"
import "strconv"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: cat [OPTION]... [FILE]...\nConcatenate FILE(s) to standard output.\n\nWith no FILE, or when FILE is -, read standard input.\n\n  -A, --show-all           equivalent to -vET\n  -b, --number-nonblank    number nonempty output lines, overrides -n\n  -e                       equivalent to -vE\n  -E, --show-ends          display $ at end of each line\n  -n, --number             number all output lines\n  -s, --squeeze-blank      suppress repeated empty output lines\n  -t                       equivalent to -vT\n  -T, --show-tabs          display TAB characters as ^I\n  -u                       (ignored)\n  -v, --show-nonprinting   use ^ notation for control characters, except LFD and TAB\n      --help               display this help and exit\n"

// copyOperand writes one operand to standard output byte for byte and reports
// whether it could be read. Standard input is passed on as it arrives.
func copyOperand(operand string) bool {
	if operand != "-" {
		chunks, total, ok := cli.ReadBytes("cat", operand, 0-1)
		cli.WriteBytes(chunks, 0, total)
		return ok
	}
	buffer := make([]byte, 65536)
	for {
		count, err := os.Stdin.Read(buffer)
		if count > 0 {
			os.Stdout.Write(buffer[:count])
		}
		if err != nil {
			return true
		}
	}
}

func main() {
	flags, values, operands := cli.ParseOptions("cat", usage, os.Args[1:], "AbeEnstTuv", "show-all=A number-nonblank=b show-ends=E number=n squeeze-blank=s show-tabs=T show-nonprinting=v", cli.ModePlain, 1)
	number := false
	nonblank := false
	squeeze := false
	ends := false
	tabs := false
	visible := false
	for index := 0; index < len(flags); index++ {
		flag := flags[index] + values[index]
		if flag == "A" {
			visible = true
			ends = true
			tabs = true
		} else if flag == "b" {
			nonblank = true
		} else if flag == "e" {
			visible = true
			ends = true
		} else if flag == "E" {
			ends = true
		} else if flag == "n" {
			number = true
		} else if flag == "s" {
			squeeze = true
		} else if flag == "t" {
			visible = true
			tabs = true
		} else if flag == "T" {
			tabs = true
		} else if flag == "v" {
			visible = true
		}
	}

	result := 0
	if !number && !nonblank && !squeeze && !ends && !tabs && !visible {
		if len(operands) == 0 {
			operands = []string{"-"}
		}
		for index := 0; index < len(operands); index++ {
			if !copyOperand(operands[index]) {
				result = 1
			}
		}
		os.Exit(result)
	}

	data, status := cli.ReadAll("cat", operands)
	result = status
	if squeeze {
		data = "\n" + data
		for strings.Contains(data, "\n\n\n") {
			data = cli.ReplaceAll(data, "\n\n\n", "\n\n")
		}
		data = data[1:]
	}
	if tabs {
		data = cli.ReplaceAll(data, "\t", "^I")
	}
	if visible {
		for code := 0; code < 32; code++ {
			if code != 9 && code != 10 {
				data = cli.ReplaceAll(data, cli.Char(code), "^"+cli.Char(code+64))
			}
		}
		data = cli.ReplaceAll(data, cli.Char(127), "^?")
	}
	if !number && !nonblank {
		if ends {
			data = cli.ReplaceAll(data, "\n", "$\n")
		}
		fmt.Print(data)
		os.Exit(result)
	}

	terminated := strings.HasSuffix(data, "\n")
	lines := cli.SplitLines(data)
	blanks := "      "
	count := 0
	for index := 0; index < len(lines); index++ {
		if nonblank && lines[index] == "" {
			continue
		}
		count++
		number := strconv.Itoa(count)
		if len(number) < 6 {
			number = blanks[len(number):] + number
		}
		lines[index] = number + "\t" + lines[index]
	}
	end := "\n"
	if ends {
		end = "$\n"
	}
	text := strings.Join(lines, end)
	if terminated {
		text = text + end
	}
	fmt.Print(text)
	os.Exit(result)
}
