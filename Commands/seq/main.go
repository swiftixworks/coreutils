package main

import "fmt"
import "os"
import "strconv"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: seq [OPTION]... LAST\n  or:  seq [OPTION]... FIRST LAST\n  or:  seq [OPTION]... FIRST INCREMENT LAST\nPrint numbers from FIRST to LAST, in steps of INCREMENT.\n\n  -s, --separator=STRING   use STRING to separate numbers (default: \\n)\n  -w, --equal-width        equalize width by padding with leading zeroes\n      --help               display this help and exit\n\nFIRST and INCREMENT default to 1. Numbers may be negative and may carry a\ndecimal fraction; the output uses the largest number of decimals given.\n"

// blockCount is how many numbers are gathered into one write.
const blockCount = 1024

// parseNumber reads a decimal number as its digits without the point and the
// count of digits after the point: "-1.50" is (-150, 2).
func parseNumber(text string) (int, int) {
	body := text
	negative := false
	if strings.HasPrefix(body, "-") {
		negative = true
		body = body[1:]
	} else if strings.HasPrefix(body, "+") {
		body = body[1:]
	}
	decimals := 0
	point := strings.Index(body, ".")
	if point >= 0 {
		decimals = len(body) - point - 1
		body = body[:point] + body[point+1:]
	}
	if !cli.IsDigits(body) || len(body) > 15 || decimals > 9 {
		cli.Fail("seq", "invalid floating point argument: '"+text+"'", 1)
	}
	value, _ := cli.ParseCount(body)
	if negative {
		value = 0 - value
	}
	return value, decimals
}

func scaled(value int, decimals int, target int) int {
	for decimals < target {
		value = value * 10
		decimals++
	}
	return value
}

// format renders a scaled value with the given number of decimals, zero-padded
// to width when width is positive.
func format(value int, decimals int, width int) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = 0 - value
	}
	unit := 1
	for index := 0; index < decimals; index++ {
		unit = unit * 10
	}
	text := strconv.Itoa(value / unit)
	if decimals > 0 {
		fraction := strconv.Itoa(value % unit)
		text = text + "." + strings.Repeat("0", decimals-len(fraction)) + fraction
	}
	missing := width - len(sign) - len(text)
	if missing > 0 {
		text = strings.Repeat("0", missing) + text
	}
	return sign + text
}

func main() {
	flags, values, operands := cli.ParseOptions("seq", usage, os.Args[1:], "s:w", "separator=s equal-width=w", cli.ModeNumbers, 1)
	separator := "\n"
	equalWidth := false
	for index := 0; index < len(flags); index++ {
		if flags[index] == "s" {
			separator = values[index]
		} else {
			equalWidth = true
		}
	}
	if len(operands) == 0 {
		cli.FailUsage("seq", "missing operand", 1)
	}
	if len(operands) > 3 {
		cli.FailUsage("seq", "extra operand '"+operands[3]+"'", 1)
	}
	first := 1
	firstDecimals := 0
	step := 1
	stepDecimals := 0
	last, lastDecimals := parseNumber(operands[len(operands)-1])
	if len(operands) >= 2 {
		value, decimals := parseNumber(operands[0])
		first = value
		firstDecimals = decimals
	}
	if len(operands) == 3 {
		value, decimals := parseNumber(operands[1])
		step = value
		stepDecimals = decimals
	}
	if step == 0 {
		cli.Fail("seq", "invalid Zero increment value: '"+operands[1]+"'", 1)
	}
	decimals := firstDecimals
	if stepDecimals > decimals {
		decimals = stepDecimals
	}
	// GNU prints with the precision of FIRST and INCREMENT; LAST only bounds.
	bound := decimals
	if lastDecimals > bound {
		bound = lastDecimals
	}
	first = scaled(first, firstDecimals, bound)
	step = scaled(step, stepDecimals, bound)
	last = scaled(last, lastDecimals, bound)
	trim := 1
	for index := decimals; index < bound; index++ {
		trim = trim * 10
	}
	if (step > 0 && first > last) || (step < 0 && first < last) {
		os.Exit(0)
	}
	count := (last-first)/step + 1
	final := first + (count-1)*step
	width := 0
	if equalWidth {
		width = len(format(first/trim, decimals, 0))
		other := len(format(final/trim, decimals, 0))
		if other > width {
			width = other
		}
	}

	plain := bound == 0 && width == 0
	block := []string{}
	value := first
	for index := 1; index < count; index++ {
		if plain {
			block = append(block, strconv.Itoa(value))
		} else {
			block = append(block, format(value/trim, decimals, width))
		}
		value = value + step
		if len(block) == blockCount {
			fmt.Print(strings.Join(block, separator) + separator)
			block = []string{}
		}
	}
	block = append(block, format(value/trim, decimals, width))
	fmt.Print(strings.Join(block, separator) + "\n")
	os.Exit(0)
}
