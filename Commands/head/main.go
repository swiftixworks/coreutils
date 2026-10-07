package main

import "fmt"
import "os"
import "strings"
import "swiftix/userland"
import "swiftix/internal/cli"

const usage = "Usage: head [OPTION]... [FILE]...\nPrint the first 10 lines of each FILE to standard output.\nWith more than one FILE, precede each with a header giving the file name.\n\nWith no FILE, or when FILE is -, read standard input.\n\n  -c, --bytes=[-]NUM       print the first NUM bytes of each file;\n                             with the leading '-', print all but the last NUM bytes\n  -n, --lines=[-]NUM       print the first NUM lines instead of the first 10;\n                             with the leading '-', print all but the last NUM lines\n  -NUM                     same as -n NUM\n  -q, --quiet, --silent    never print headers giving file names\n  -v, --verbose            always print headers giving file names\n      --help               display this help and exit\n\nNUM may have a multiplier suffix: b 512, kB 1000, K 1024, MB 1000000, M 1048576.\n"

// readStdinLines reads standard input only until it holds count lines, so an
// endless producer is not drained.
func readStdinLines(count int) string {
	parts := []string{}
	seen := 0
	for seen < count {
		chunk, status := userland.ReadStdin()
		if status != 0 {
			break
		}
		parts = append(parts, chunk)
		seen = seen + strings.Count(chunk, "\n")
	}
	return strings.Join(parts, "")
}

func main() {
	flags, values, operands := cli.ParseOptions("head", usage, os.Args[1:], "n:c:qv", "lines=n bytes=c quiet=q silent=q verbose=v", cli.ModeCount, 1)
	bytes := false
	fromEnd := false
	count := 10
	quiet := false
	verbose := false
	for index := 0; index < len(flags); index++ {
		flag := flags[index]
		value := values[index]
		if flag == "q" {
			quiet = true
			verbose = false
		} else if flag == "v" {
			verbose = true
			quiet = false
		} else {
			bytes = flag == "c"
			fromEnd = false
			if flag != "#" && strings.HasPrefix(value, "-") {
				fromEnd = true
				value = value[1:]
			} else if strings.HasPrefix(value, "+") {
				value = value[1:]
			}
			parsed, ok := cli.ParseSize(value)
			if !ok {
				unit := "lines"
				if bytes {
					unit = "bytes"
				}
				cli.Fail("head", "invalid number of "+unit+": '"+values[index]+"'", 1)
			}
			count = parsed
		}
	}
	if len(operands) == 0 {
		operands = []string{"-"}
	}
	headers := verbose || (len(operands) > 1 && !quiet)
	result := 0
	printed := false
	for index := 0; index < len(operands); index++ {
		operand := operands[index]
		name := operand
		if operand == "-" {
			name = "standard input"
		}
		header := ""
		if headers {
			header = "==> " + name + " <==\n"
			if printed {
				header = "\n" + header
			}
		}
		if bytes {
			// Byte counts are exact: the input is never decoded as text.
			limit := count
			if fromEnd {
				limit = 0 - 1
			}
			chunks, total, ok := cli.ReadBytes("head", operand, limit)
			if !ok {
				result = 1
				continue
			}
			fmt.Print(header)
			printed = true
			end := count
			if fromEnd {
				end = total - count
			}
			cli.WriteBytes(chunks, 0, end)
			continue
		}
		data := ""
		if operand != "-" {
			contents, status := userland.ReadInput("head", []string{operand})
			if status != 0 {
				result = 1
				continue
			}
			data = contents
		} else if fromEnd {
			data = cli.ReadStdin()
		} else {
			data = readStdinLines(count)
		}
		fmt.Print(header)
		printed = true
		if fromEnd {
			fmt.Print(data[:cli.LineOffset(data, cli.LineCount(data)-count)])
		} else {
			fmt.Print(data[:cli.LineOffset(data, count)])
		}
	}
	os.Exit(result)
}
