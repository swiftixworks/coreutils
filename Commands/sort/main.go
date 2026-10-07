package main

import "fmt"
import "os"
import "sort"
import "strconv"
import "strings"
import "swiftix/userland"
import "swiftix/internal/cli"

const usage = "Usage: sort [OPTION]... [FILE]...\nWrite sorted concatenation of all FILE(s) to standard output.\n\nWith no FILE, or when FILE is -, read standard input.\n\nOrdering options:\n  -b, --ignore-leading-blanks   ignore leading blanks\n  -f, --ignore-case             fold lower case to upper case characters\n  -h, --human-numeric-sort      compare human readable numbers (e.g., 2K 1G)\n  -n, --numeric-sort            compare according to string numerical value\n  -r, --reverse                 reverse the result of comparisons\n\nOther options:\n  -c, --check                   check for sorted input; do not sort\n  -C                            like -c, but do not report first bad line\n  -k, --key=KEYDEF              sort via a key; KEYDEF gives location and type\n  -m, --merge                   merge already sorted files (sorts them)\n  -o, --output=FILE             write result to FILE instead of standard output\n  -s, --stable                  stabilize sort by disabling last-resort comparison\n  -t, --field-separator=SEP     use SEP instead of non-blank to blank transition\n  -u, --unique                  output only the first of an equal run\n      --help                    display this help and exit\n\nKEYDEF is F[.C][OPTS][,F[.C][OPTS]] for start and stop position, where F is a\nfield number and C a character position in the field; both are origin 1, and\nthe stop position defaults to the line's end. OPTS is one or more of the\nordering options bfhnr, which override global ordering options for that key.\nKeys may not mix ascending and descending order.\n\nComparison is by byte value (the C locale).\n"

// One sort key: the field range and the ordering applied to it.
type sortKey struct {
	startField  int
	startChar   int
	endField    int
	endChar     int
	blanksStart bool
	blanksEnd   bool
	fold        bool
	numeric     bool
	human       bool
	reverse     bool
	custom      bool
}

var separator = ""

func newKey() sortKey {
	return sortKey{startField: 1, startChar: 1, endField: 0, endChar: 0, blanksStart: false, blanksEnd: false, fold: false, numeric: false, human: false, reverse: false, custom: false}
}

func isBlank(code byte) bool {
	return code == 32 || code == 9
}

// parsePosition reads "F[.C][OPTS]" and returns field, char, and the option
// letters that followed.
func parsePosition(text string, whole string) (int, int, string) {
	end := 0
	for end < len(text) && text[end] >= 48 && text[end] <= 57 {
		end++
	}
	field, ok := cli.ParseCount(text[:end])
	if !ok {
		cli.Fail("sort", "invalid number at field start: invalid count at start of '"+whole+"'", 2)
	}
	character := 0
	rest := text[end:]
	if strings.HasPrefix(rest, ".") {
		rest = rest[1:]
		end = 0
		for end < len(rest) && rest[end] >= 48 && rest[end] <= 57 {
			end++
		}
		value, valid := cli.ParseCount(rest[:end])
		if !valid {
			cli.Fail("sort", "invalid number after '.': invalid count at start of '"+rest+"'", 2)
		}
		character = value
		rest = rest[end:]
	}
	return field, character, rest
}

func applyOrdering(key sortKey, letters string, start bool, whole string) sortKey {
	for index := 0; index < len(letters); index++ {
		letter := letters[index]
		key.custom = true
		if letter == 98 {
			if start {
				key.blanksStart = true
			} else {
				key.blanksEnd = true
			}
		} else if letter == 102 {
			key.fold = true
		} else if letter == 110 {
			key.numeric = true
		} else if letter == 104 {
			key.human = true
		} else if letter == 114 {
			key.reverse = true
		} else {
			cli.Fail("sort", "stray character in field spec: invalid field specification '"+whole+"'", 2)
		}
	}
	return key
}

func parseKey(text string) sortKey {
	key := newKey()
	startText := text
	endText := ""
	comma := strings.Index(text, ",")
	if comma >= 0 {
		startText = text[:comma]
		endText = text[comma+1:]
	}
	field, character, letters := parsePosition(startText, text)
	if field == 0 {
		cli.Fail("sort", "-: invalid field specification '"+text+"'\nfield number is zero", 2)
	}
	if character == 0 {
		character = 1
	}
	key.startField = field
	key.startChar = character
	key = applyOrdering(key, letters, true, text)
	if comma >= 0 {
		endField, endChar, endLetters := parsePosition(endText, text)
		if endField == 0 {
			cli.Fail("sort", "-: invalid field specification '"+text+"'\nfield number is zero", 2)
		}
		key.endField = endField
		key.endChar = endChar
		key = applyOrdering(key, endLetters, false, text)
	}
	return key
}

// fieldBounds returns the byte offsets where field number (origin 1) starts and
// ends in line. Without -t a field is its leading blanks plus the non-blanks.
func fieldBounds(line string, number int) (int, int) {
	position := 0
	start := 0
	if separator != "" {
		for field := 1; field <= number; field++ {
			start = position
			next := strings.Index(line[position:], separator)
			if next < 0 {
				if field < number {
					return len(line), len(line)
				}
				return start, len(line)
			}
			if field == number {
				return start, position + next
			}
			position = position + next + len(separator)
		}
		return start, len(line)
	}
	for field := 1; field <= number; field++ {
		start = position
		for position < len(line) && isBlank(line[position]) {
			position++
		}
		for position < len(line) && !isBlank(line[position]) {
			position++
		}
	}
	return start, position
}

// keyText extracts the part of line a key covers.
func keyText(line string, key sortKey) string {
	begin, beginEnd := fieldBounds(line, key.startField)
	if key.blanksStart {
		for begin < beginEnd && isBlank(line[begin]) {
			begin++
		}
	}
	begin = begin + key.startChar - 1
	if begin > len(line) {
		begin = len(line)
	}
	end := len(line)
	if key.endField > 0 {
		fieldStart, fieldEnd := fieldBounds(line, key.endField)
		if key.endChar == 0 {
			end = fieldEnd
		} else {
			if key.blanksEnd {
				for fieldStart < fieldEnd && isBlank(line[fieldStart]) {
					fieldStart++
				}
			}
			end = fieldStart + key.endChar
			if end > len(line) {
				end = len(line)
			}
		}
	}
	if end < begin {
		return ""
	}
	return line[begin:end]
}

// lengthLetters[n] is the character that stands for an integer part of n
// digits; the characters ascend with n.
const lengthLetters = "0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijkl"

// numericKey encodes the number that starts text so that byte order equals
// numeric order. Zero and non-numbers encode as "0". A positive number is the
// length character of its integer part, the integer digits, and the fraction,
// so a longer integer sorts later and fractions extend their integer. A
// negative number is "-" followed by the complement of all that and a high
// terminator. With human set, a rank for the K, M, G... suffix comes first.
func numericKey(text string, human bool) string {
	position := 0
	for position < len(text) && isBlank(text[position]) {
		position++
	}
	negative := false
	if position < len(text) && text[position] == 45 {
		negative = true
		position++
	}
	for position < len(text) && text[position] == 48 {
		position++
	}
	start := position
	for position < len(text) && text[position] >= 48 && text[position] <= 57 {
		position++
	}
	whole := text[start:position]
	fraction := ""
	if position < len(text) && text[position] == 46 {
		position++
		start = position
		for position < len(text) && text[position] >= 48 && text[position] <= 57 {
			position++
		}
		fraction = text[start:position]
		for strings.HasSuffix(fraction, "0") {
			fraction = fraction[:len(fraction)-1]
		}
	}
	if whole == "" && fraction == "" {
		return "0"
	}
	length := len(whole)
	if length > 59 {
		length = 59
	}
	rank := ""
	flippedRank := ""
	if human {
		order := 0
		if position < len(text) {
			order = strings.Index("KMGTPEZY", text[position:position+1]) + 1
			if order == 0 && text[position] == 107 {
				order = 1
			}
		}
		rank = cli.Char(48 + order)
		flippedRank = cli.Char(57 - order)
	}
	if !negative {
		return rank + lengthLetters[length:length+1] + whole + fraction
	}
	digits := whole + fraction
	flipped := ""
	for index := 0; index < len(digits); index++ {
		flipped = flipped + cli.Char(105-int(digits[index]))
	}
	return "-" + flippedRank + cli.Char(110-length) + flipped + "~"
}

func stripBlanks(text string) string {
	position := 0
	for position < len(text) && isBlank(text[position]) {
		position++
	}
	return text[position:]
}

func encodeKey(text string, key sortKey) string {
	if key.human {
		return numericKey(text, true)
	}
	if key.numeric {
		return numericKey(text, false)
	}
	if key.blanksStart && key.startChar == 1 && key.startField == 1 && key.endField == 0 {
		text = stripBlanks(text)
	}
	return text
}

func main() {
	flags, values, operands := cli.ParseOptions("sort", usage, os.Args[1:], "bfhnrcCk:mo:st:u", "ignore-leading-blanks=b ignore-case=f human-numeric-sort=h numeric-sort=n reverse=r check=c key=k merge=m output=o stable=s field-separator=t unique=u", cli.ModePlain, 2)
	global := newKey()
	keys := []sortKey{}
	check := false
	quietCheck := false
	stable := false
	unique := false
	output := ""
	for index := 0; index < len(flags); index++ {
		flag := flags[index]
		value := values[index]
		if flag == "b" {
			global.blanksStart = true
			global.blanksEnd = true
		} else if flag == "f" {
			global.fold = true
		} else if flag == "h" {
			global.human = true
		} else if flag == "n" {
			global.numeric = true
		} else if flag == "r" {
			global.reverse = true
		} else if flag == "c" {
			check = true
		} else if flag == "C" {
			check = true
			quietCheck = true
		} else if flag == "k" {
			keys = append(keys, parseKey(value))
		} else if flag == "o" {
			output = value
		} else if flag == "s" {
			stable = true
		} else if flag == "t" {
			if value == "\\0" {
				value = cli.Char(0)
			}
			if value == "" {
				cli.Fail("sort", "empty tab", 2)
			}
			if cli.CharCount(value) > 1 {
				cli.Fail("sort", "multi-character tab '"+value+"'", 2)
			}
			separator = value
		} else if flag == "u" {
			unique = true
		}
	}
	// Keys without their own ordering letters inherit the global ones.
	for index := 0; index < len(keys); index++ {
		key := keys[index]
		if !key.custom {
			key.blanksStart = global.blanksStart
			key.blanksEnd = global.blanksEnd
			key.fold = global.fold
			key.numeric = global.numeric
			key.human = global.human
			key.reverse = global.reverse
			keys[index] = key
		}
	}
	plain := len(keys) == 0 && !global.blanksStart && !global.fold && !global.numeric && !global.human
	if len(keys) == 0 {
		keys = append(keys, global)
	}
	reverse := keys[0].reverse
	for index := 1; index < len(keys); index++ {
		if keys[index].reverse != reverse {
			cli.Fail("sort", "keys that mix ascending and descending order are not supported", 2)
		}
	}

	data, status := cli.ReadAll("sort", operands)
	if status != 0 {
		os.Exit(2)
	}
	lines := cli.SplitLines(data)
	total := len(lines)
	nul := cli.Char(0)

	// Build one comparison string per line whose byte order is the requested
	// order: the encoded keys separated by NUL, which sorts below every other
	// byte, so a shorter key precedes its extensions.
	keyed := lines
	if !plain {
		if strings.Contains(data, nul) {
			cli.Fail("sort", "input contains NUL bytes, which key comparison does not support", 2)
		}
		folded := lines
		anyFold := false
		for index := 0; index < len(keys); index++ {
			if keys[index].fold {
				anyFold = true
			}
		}
		if anyFold {
			folded = cli.SplitLines(cli.UpperASCII(data))
		}
		first := keys[0]
		wholeLine := len(keys) == 1 && first.startField == 1 && first.startChar == 1 && first.endField == 0
		if wholeLine && !first.numeric && !first.human && !first.blanksStart {
			keyed = folded
		} else if wholeLine && first.numeric && !first.human {
			// "sort -n". A line that is a plain positive integer is recognised
			// and encoded in place, because a helper call per line is costly
			// on a large input; any other line takes the general scan.
			keyed = make([]string, total)
			for index := 0; index < total; index++ {
				line := lines[index]
				size := len(line)
				if size > 0 && size <= 18 && line[0] >= 49 && line[0] <= 57 {
					_, err := strconv.Atoi(line)
					if err == nil {
						keyed[index] = lengthLetters[size:size+1] + line
						continue
					}
				}
				keyed[index] = numericKey(line, false)
			}
		} else if wholeLine {
			keyed = make([]string, total)
			for index := 0; index < total; index++ {
				keyed[index] = encodeKey(folded[index], first)
			}
		} else {
			keyed = make([]string, total)
			for index := 0; index < total; index++ {
				combined := ""
				for position := 0; position < len(keys); position++ {
					key := keys[position]
					source := lines[index]
					if key.fold {
						source = folded[index]
					}
					if position > 0 {
						combined = combined + nul
					}
					combined = combined + encodeKey(keyText(source, key), key)
				}
				keyed[index] = combined
			}
		}
	}

	if check {
		for index := 1; index < total; index++ {
			order := 0
			if keyed[index] < keyed[index-1] {
				order = 0 - 1
			} else if keyed[index-1] < keyed[index] {
				order = 1
			} else if !stable && !unique && !plain {
				if lines[index] < lines[index-1] {
					order = 0 - 1
				} else if lines[index-1] < lines[index] {
					order = 1
				}
			}
			if reverse {
				order = 0 - order
			}
			if order < 0 || (unique && order == 0) {
				if !quietCheck {
					name := "-"
					if len(operands) > 0 {
						name = operands[0]
					}
					cli.Fail("sort", name+":"+strconv.Itoa(index+1)+": disorder: "+lines[index], 1)
				}
				os.Exit(1)
			}
		}
		os.Exit(0)
	}

	// Decorate. A plain sort orders the lines themselves. Otherwise an item is
	// the key and then either the line, as the last-resort comparison that also
	// carries the text to print, or (for a stable or unique sort) a fixed-width
	// sequence number that keeps equal keys in input order and names the line.
	// Items are sorted ascending and read backwards for a reversed sort, so a
	// reversed sort numbers the lines backwards to keep ties in input order.
	tieByLine := !stable && !unique
	items := lines
	width := cli.DigitCount(total)
	if !plain && tieByLine {
		items = make([]string, total)
		for index := 0; index < total; index++ {
			items[index] = keyed[index] + nul + lines[index]
		}
	} else if !plain {
		items = make([]string, total)
		zeros := strings.Repeat("0", width)
		for index := 0; index < total; index++ {
			order := index
			if reverse {
				order = total - 1 - index
			}
			number := strconv.Itoa(order)
			items[index] = keyed[index] + nul + zeros[len(number):] + number
		}
	}

	sort.Strings(items)

	// Undecorate in output order, dropping repeats of a key for -u.
	sorted := items
	if !plain || unique || reverse {
		sorted = make([]string, 0)
		position := 0
		step := 1
		if reverse {
			position = total - 1
			step = 0 - 1
		}
		previous := ""
		for seen := 0; seen < total; seen++ {
			item := items[position]
			position = position + step
			if plain {
				if unique && seen > 0 && item == previous {
					continue
				}
				previous = item
				sorted = append(sorted, item)
			} else if tieByLine {
				sorted = append(sorted, item[strings.LastIndex(item, nul)+1:])
			} else {
				order, _ := strconv.Atoi(item[len(item)-width:])
				if reverse {
					order = total - 1 - order
				}
				if unique {
					if seen > 0 && keyed[order] == previous {
						continue
					}
					previous = keyed[order]
				}
				sorted = append(sorted, lines[order])
			}
		}
	}
	result := ""
	if len(sorted) > 0 {
		result = strings.Join(sorted, "\n") + "\n"
	}
	if output == "" {
		fmt.Print(result)
	} else if userland.WriteFile(output, result) != 0 {
		cli.Fail("sort", "open failed: "+output, 2)
	}
	os.Exit(0)
}
