package main

import "fmt"
import "os"
import "strings"
import "time"
import "swiftix/internal/cli"

const usage = "Usage: tail [OPTION]... [FILE]...\nPrint the last 10 lines of each FILE to standard output.\nWith more than one FILE, precede each with a header giving the file name.\n\nWith no FILE, or when FILE is -, read standard input.\n\n  -c, --bytes=[+]NUM       output the last NUM bytes; or use -c +NUM to\n                             output starting with byte NUM of each file\n  -f, -F, --follow         output appended data as the file grows\n  -n, --lines=[+]NUM       output the last NUM lines, instead of the last 10;\n                             or use -n +NUM to output starting with line NUM\n  -NUM                     same as -n NUM\n  -q, --quiet, --silent    never output headers giving file names\n  -s, --sleep-interval=N   with -f, sleep for approximately N seconds\n                             (default 1.0) between looks at the files\n  -v, --verbose            always output headers giving file names\n      --help               display this help and exit\n\nNUM may have a multiplier suffix: b 512, kB 1000, K 1024, MB 1000000, M 1048576.\n\nWith -f, tail keeps running after the end of a named file and prints what is\nappended to it; a file that shrinks is printed again from its start. Standard\ninput is read to its end and not followed further.\n"

// tailOffset returns the byte offset where the last count lines of data begin.
func tailOffset(data string, count int) int {
	if count <= 0 {
		return len(data)
	}
	body := data
	if strings.HasSuffix(body, "\n") {
		body = body[:len(body)-1]
	}
	if count > strings.Count(body, "\n") {
		return 0
	}
	if count <= 24 {
		rest := body
		for index := 0; index < count; index++ {
			rest = rest[:strings.LastIndex(rest, "\n")]
		}
		return len(rest) + 1
	}
	lines := strings.Split(body, "\n")
	return len(body) - len(strings.Join(lines[len(lines)-count:], "\n"))
}

// parseInterval reads a -s value, seconds with an optional decimal fraction,
// as milliseconds.
func parseInterval(text string) (int, bool) {
	whole := text
	fraction := ""
	point := strings.Index(text, ".")
	if point >= 0 {
		whole = text[:point]
		fraction = text[point+1:]
	}
	if whole == "" && fraction != "" {
		whole = "0"
	}
	seconds, ok := cli.ParseCount(whole)
	if !ok || (fraction != "" && !cli.IsDigits(fraction)) {
		return 0, false
	}
	if seconds > 1000000000 {
		seconds = 1000000000
	}
	thousandths, _ := cli.ParseCount((fraction + "000")[:3])
	return seconds*1000 + thousandths, true
}

// pause sleeps for the given number of milliseconds. The process is parked,
// so other processes run meanwhile.
func pause(milliseconds int) {
	for milliseconds >= 1000 {
		time.Sleep(time.Second)
		milliseconds = milliseconds - 1000
	}
	for milliseconds > 0 {
		time.Sleep(time.Millisecond)
		milliseconds--
	}
}

func main() {
	flags, values, operands := cli.ParseOptions("tail", usage, os.Args[1:], "n:c:fFqvs:", "lines=n bytes=c follow=f quiet=q silent=q verbose=v sleep-interval=s", cli.ModeCount, 1)
	bytes := false
	fromStart := false
	count := 10
	quiet := false
	verbose := false
	follow := false
	interval := 1000
	for index := 0; index < len(flags); index++ {
		flag := flags[index]
		value := values[index]
		if flag == "q" {
			quiet = true
			verbose = false
		} else if flag == "v" {
			verbose = true
			quiet = false
		} else if flag == "f" || flag == "F" {
			follow = true
		} else if flag == "s" {
			parsed, ok := parseInterval(value)
			if !ok {
				cli.Fail("tail", "invalid number of seconds: '"+value+"'", 1)
			}
			interval = parsed
		} else {
			bytes = flag == "c"
			fromStart = false
			if flag != "#" && strings.HasPrefix(value, "+") {
				fromStart = true
				value = value[1:]
			} else if strings.HasPrefix(value, "-") {
				value = value[1:]
			}
			parsed, ok := cli.ParseSize(value)
			if !ok {
				unit := "lines"
				if bytes {
					unit = "bytes"
				}
				cli.Fail("tail", "invalid number of "+unit+": '"+values[index]+"'", 1)
			}
			count = parsed
		}
	}
	if len(operands) == 0 {
		operands = []string{"-"}
	}
	headers := verbose || (len(operands) > 1 && !quiet)
	result := 0
	// sizes[index] is how many bytes of a named file have been dealt with, or
	// -1 for an operand that is not followed. current is the operand whose
	// header was printed last.
	sizes := make([]int, len(operands))
	following := 0
	current := 0 - 1
	for index := 0; index < len(operands); index++ {
		operand := operands[index]
		name := operand
		if operand == "-" {
			name = "standard input"
		}
		sizes[index] = 0 - 1
		data := ""
		chunks := [][]byte{}
		total := 0
		if operand == "-" && !bytes {
			data = cli.ReadStdin()
		} else {
			contents, length, ok := cli.ReadBytes("tail", operand, 0-1)
			if !ok {
				result = 1
				continue
			}
			chunks = contents
			total = length
			if operand != "-" {
				sizes[index] = total
				following++
			}
			if !bytes {
				data = string(chunks[0])
			}
		}
		if headers {
			if current >= 0 {
				fmt.Print("\n")
			}
			fmt.Print("==> " + name + " <==\n")
		}
		current = index
		if bytes {
			// Byte counts are exact: the input is never decoded as text.
			start := total - count
			if fromStart {
				start = count - 1
			}
			cli.WriteBytes(chunks, start, total)
		} else if fromStart {
			fmt.Print(data[cli.LineOffset(data, count-1):])
		} else {
			fmt.Print(data[tailOffset(data, count):])
		}
	}
	if !follow || following == 0 {
		os.Exit(result)
	}

	// Follow the named files by looking at them again after every interval.
	// The loop ends when the process is interrupted or its output is closed.
	if interval < 1 {
		interval = 1
	}
	for {
		pause(interval)
		for index := 0; index < len(operands); index++ {
			if sizes[index] < 0 {
				continue
			}
			data, err := os.ReadFile(operands[index])
			if err != nil {
				// Gone for now; if it comes back it is printed from its start.
				sizes[index] = 0
				continue
			}
			if len(data) < sizes[index] {
				fmt.Fprint(os.Stderr, "tail: "+operands[index]+": file truncated\n")
				sizes[index] = 0
			}
			if len(data) == sizes[index] {
				continue
			}
			if headers && current != index {
				fmt.Print("\n==> " + operands[index] + " <==\n")
			}
			current = index
			os.Stdout.Write(data[sizes[index]:])
			sizes[index] = len(data)
		}
	}
}
