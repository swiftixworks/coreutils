package main

import "fmt"
import "os"
import "strings"
import "swiftix/internal/cli"

// escapeValue reads up to limit digits of the given base from text and returns
// the value and the number of bytes consumed.
func escapeValue(text string, base int, limit int) (int, int) {
	value := 0
	used := 0
	for used < limit && used < len(text) {
		code := int(text[used])
		digit := 0 - 1
		if code >= 48 && code <= 57 {
			digit = code - 48
		} else if code >= 97 && code <= 102 {
			digit = code - 87
		} else if code >= 65 && code <= 70 {
			digit = code - 55
		}
		if digit < 0 || digit >= base {
			break
		}
		value = value*base + digit
		used++
	}
	return value, used
}

// unescape expands backslash escapes and reports whether \c ended the output.
func unescape(text string) (string, bool) {
	pieces := strings.Split(text, "\\")
	result := pieces[0]
	literal := false
	for index := 1; index < len(pieces); index++ {
		piece := pieces[index]
		if literal {
			literal = false
			result = result + piece
			continue
		}
		if piece == "" {
			result = result + "\\"
			literal = true
			continue
		}
		code := piece[0]
		rest := piece[1:]
		if code == 97 {
			result = result + cli.Char(7) + rest
		} else if code == 98 {
			result = result + cli.Char(8) + rest
		} else if code == 99 {
			return result, true
		} else if code == 101 {
			result = result + cli.Char(27) + rest
		} else if code == 102 {
			result = result + cli.Char(12) + rest
		} else if code == 110 {
			result = result + "\n" + rest
		} else if code == 114 {
			result = result + "\r" + rest
		} else if code == 116 {
			result = result + "\t" + rest
		} else if code == 118 {
			result = result + cli.Char(11) + rest
		} else if code == 48 {
			value, used := escapeValue(rest, 8, 3)
			result = result + cli.Char(value%256) + rest[used:]
		} else if code == 120 {
			value, used := escapeValue(rest, 16, 2)
			if used == 0 {
				result = result + "\\" + piece
			} else {
				result = result + cli.Char(value) + rest[used:]
			}
		} else {
			result = result + "\\" + piece
		}
	}
	return result, false
}

func main() {
	args := os.Args[1:]
	newline := true
	escapes := false
	first := 0
	for first < len(args) {
		arg := args[first]
		if len(arg) < 2 || arg[0] != 45 {
			break
		}
		valid := true
		for index := 1; index < len(arg); index++ {
			if arg[index] != 110 && arg[index] != 101 && arg[index] != 69 {
				valid = false
			}
		}
		if !valid {
			break
		}
		for index := 1; index < len(arg); index++ {
			if arg[index] == 110 {
				newline = false
			} else if arg[index] == 101 {
				escapes = true
			} else {
				escapes = false
			}
		}
		first++
	}
	text := strings.Join(args[first:], " ")
	if escapes && strings.Contains(text, "\\") {
		expanded, stop := unescape(text)
		text = expanded
		if stop {
			newline = false
		}
	}
	if newline {
		text = text + "\n"
	}
	fmt.Print(text)
	os.Exit(0)
}
