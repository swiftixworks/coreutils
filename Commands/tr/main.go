package main

import "fmt"
import "os"
import "strings"
import "swiftix/internal/cli"

const usage = "Usage: tr [OPTION]... SET1 [SET2]\nTranslate, squeeze, and/or delete characters from standard input,\nwriting to standard output.\n\n  -c, -C, --complement    use the complement of SET1\n  -d, --delete            delete characters in SET1, do not translate\n  -s, --squeeze-repeats   replace each sequence of a repeated character\n                            that is listed in the last specified SET,\n                            with a single occurrence of that character\n  -t, --truncate-set1     first truncate SET1 to length of SET2\n      --help              display this help and exit\n\nSETs are strings of characters. Interpreted sequences are:\n\n  \\NNN            character with octal value NNN (1 to 3 octal digits)\n  \\\\ \\a \\b \\f \\n \\r \\t \\v   backslash and the usual control characters\n  CHAR1-CHAR2     all characters from CHAR1 to CHAR2 in ascending order\n  [CHAR*]         in SET2, copies of CHAR until length of SET1\n  [CHAR*REPEAT]   REPEAT copies of CHAR\n  [:alnum:] [:alpha:] [:blank:] [:cntrl:] [:digit:] [:graph:] [:lower:]\n  [:print:] [:punct:] [:space:] [:upper:] [:xdigit:]\n"

// placeholders are 256 private-use characters, three UTF-8 bytes each. A
// translation whose sets overlap goes through them so that no character is
// translated twice.
const placeholders = "\ue000\ue001\ue002\ue003\ue004\ue005\ue006\ue007\ue008\ue009\ue00a\ue00b\ue00c\ue00d\ue00e\ue00f\ue010\ue011\ue012\ue013\ue014\ue015\ue016\ue017\ue018\ue019\ue01a\ue01b\ue01c\ue01d\ue01e\ue01f\ue020\ue021\ue022\ue023\ue024\ue025\ue026\ue027\ue028\ue029\ue02a\ue02b\ue02c\ue02d\ue02e\ue02f\ue030\ue031\ue032\ue033\ue034\ue035\ue036\ue037\ue038\ue039\ue03a\ue03b\ue03c\ue03d\ue03e\ue03f\ue040\ue041\ue042\ue043\ue044\ue045\ue046\ue047\ue048\ue049\ue04a\ue04b\ue04c\ue04d\ue04e\ue04f\ue050\ue051\ue052\ue053\ue054\ue055\ue056\ue057\ue058\ue059\ue05a\ue05b\ue05c\ue05d\ue05e\ue05f\ue060\ue061\ue062\ue063\ue064\ue065\ue066\ue067\ue068\ue069\ue06a\ue06b\ue06c\ue06d\ue06e\ue06f\ue070\ue071\ue072\ue073\ue074\ue075\ue076\ue077\ue078\ue079\ue07a\ue07b\ue07c\ue07d\ue07e\ue07f\ue080\ue081\ue082\ue083\ue084\ue085\ue086\ue087\ue088\ue089\ue08a\ue08b\ue08c\ue08d\ue08e\ue08f\ue090\ue091\ue092\ue093\ue094\ue095\ue096\ue097\ue098\ue099\ue09a\ue09b\ue09c\ue09d\ue09e\ue09f\ue0a0\ue0a1\ue0a2\ue0a3\ue0a4\ue0a5\ue0a6\ue0a7\ue0a8\ue0a9\ue0aa\ue0ab\ue0ac\ue0ad\ue0ae\ue0af\ue0b0\ue0b1\ue0b2\ue0b3\ue0b4\ue0b5\ue0b6\ue0b7\ue0b8\ue0b9\ue0ba\ue0bb\ue0bc\ue0bd\ue0be\ue0bf\ue0c0\ue0c1\ue0c2\ue0c3\ue0c4\ue0c5\ue0c6\ue0c7\ue0c8\ue0c9\ue0ca\ue0cb\ue0cc\ue0cd\ue0ce\ue0cf\ue0d0\ue0d1\ue0d2\ue0d3\ue0d4\ue0d5\ue0d6\ue0d7\ue0d8\ue0d9\ue0da\ue0db\ue0dc\ue0dd\ue0de\ue0df\ue0e0\ue0e1\ue0e2\ue0e3\ue0e4\ue0e5\ue0e6\ue0e7\ue0e8\ue0e9\ue0ea\ue0eb\ue0ec\ue0ed\ue0ee\ue0ef\ue0f0\ue0f1\ue0f2\ue0f3\ue0f4\ue0f5\ue0f6\ue0f7\ue0f8\ue0f9\ue0fa\ue0fb\ue0fc\ue0fd\ue0fe\ue0ff"

// codeOf returns the code of a one-character string below 256, or -1.
func codeOf(character string) int {
	if len(character) == 1 {
		return int(character[0])
	}
	if len(character) == 2 && character[0] >= 194 && character[0] <= 195 {
		return (int(character[0])-192)*64 + int(character[1]) - 128
	}
	return 0 - 1
}

func codeRange(low int, high int) []string {
	result := []string{}
	for code := low; code <= high; code++ {
		result = append(result, cli.Char(code))
	}
	return result
}

// classMembers returns the characters of a POSIX class, or nil-length with ok
// false when the name is unknown.
func classMembers(name string) ([]string, bool) {
	result := []string{}
	known := true
	for code := 0; code < 128; code++ {
		upper := code >= 65 && code <= 90
		lower := code >= 97 && code <= 122
		digit := code >= 48 && code <= 57
		graph := code > 32 && code < 127
		member := false
		if name == "alnum" {
			member = upper || lower || digit
		} else if name == "alpha" {
			member = upper || lower
		} else if name == "blank" {
			member = code == 32 || code == 9
		} else if name == "cntrl" {
			member = code < 32 || code == 127
		} else if name == "digit" {
			member = digit
		} else if name == "graph" {
			member = graph
		} else if name == "lower" {
			member = lower
		} else if name == "print" {
			member = graph || code == 32
		} else if name == "punct" {
			member = graph && !upper && !lower && !digit
		} else if name == "space" {
			member = code == 32 || (code >= 9 && code <= 13)
		} else if name == "upper" {
			member = upper
		} else if name == "xdigit" {
			member = digit || (code >= 65 && code <= 70) || (code >= 97 && code <= 102)
		} else {
			known = false
		}
		if member {
			result = append(result, cli.Char(code))
		}
	}
	return result, known
}

// Parsed set state: the expanded characters, and where an unbounded "[c*]"
// repeat sits (-1 when absent) so SET2 can be padded there.
var setRepeatAt = 0 - 1
var setRepeatChar = ""

// expandSet turns a SET operand into its characters.
func expandSet(set string) []string {
	characters := strings.Split(set, "")
	total := len(characters)
	// Resolve escapes first; literal marks characters that came from one so
	// "\\-" and "\\[" lose their special meaning.
	items := []string{}
	literal := []bool{}
	for index := 0; index < total; index++ {
		character := characters[index]
		if character != "\\" || index+1 >= total {
			items = append(items, character)
			literal = append(literal, false)
			continue
		}
		index++
		next := characters[index]
		code := codeOf(next)
		if code >= 48 && code <= 55 {
			value := code - 48
			digits := 1
			for digits < 3 && index+1 < total {
				following := codeOf(characters[index+1])
				if following < 48 || following > 55 {
					break
				}
				value = value*8 + following - 48
				index++
				digits++
			}
			next = cli.Char(value % 256)
		} else if next == "a" {
			next = cli.Char(7)
		} else if next == "b" {
			next = cli.Char(8)
		} else if next == "f" {
			next = cli.Char(12)
		} else if next == "n" {
			next = "\n"
		} else if next == "r" {
			next = "\r"
		} else if next == "t" {
			next = "\t"
		} else if next == "v" {
			next = cli.Char(11)
		}
		items = append(items, next)
		literal = append(literal, true)
	}

	setRepeatAt = 0 - 1
	result := []string{}
	count := len(items)
	for index := 0; index < count; index++ {
		item := items[index]
		if item == "[" && !literal[index] && index+1 < count {
			// [:class:]
			if items[index+1] == ":" {
				close := 0 - 1
				for position := index + 2; position+1 < count; position++ {
					if items[position] == ":" && items[position+1] == "]" {
						close = position
						break
					}
				}
				if close > 0 {
					name := strings.Join(items[index+2:close], "")
					members, known := classMembers(name)
					if !known {
						cli.Fail("tr", "invalid character class '"+name+"'", 1)
					}
					for position := 0; position < len(members); position++ {
						result = append(result, members[position])
					}
					index = close + 1
					continue
				}
			}
			// [c*] and [c*n]
			if index+2 < count && items[index+2] == "*" {
				close := 0 - 1
				for position := index + 3; position < count; position++ {
					if items[position] == "]" {
						close = position
						break
					}
				}
				if close > 0 {
					digits := strings.Join(items[index+3:close], "")
					if digits == "" || digits == "0" {
						setRepeatAt = len(result)
						setRepeatChar = items[index+1]
						index = close
						continue
					}
					repeat, ok := cli.ParseCount(digits)
					if ok {
						if strings.HasPrefix(digits, "0") {
							repeat = 0
							for position := 0; position < len(digits); position++ {
								repeat = repeat*8 + int(digits[position]) - 48
							}
						}
						for position := 0; position < repeat; position++ {
							result = append(result, items[index+1])
						}
						index = close
						continue
					}
				}
			}
		}
		if index+2 < count && items[index+1] == "-" && !literal[index+1] {
			low := codeOf(item)
			high := codeOf(items[index+2])
			if low < 0 || high < 0 {
				cli.Fail("tr", "range '"+item+"-"+items[index+2]+"' is outside the supported characters (codes 0-255)", 1)
			}
			if low > high {
				cli.Fail("tr", "range-endpoints of '"+item+"-"+items[index+2]+"' are in reverse collating sequence order", 1)
			}
			for code := low; code <= high; code++ {
				result = append(result, cli.Char(code))
			}
			index = index + 2
			continue
		}
		result = append(result, item)
	}
	return result
}

// distinct returns the members of set without repeats, in first-seen order.
func distinct(set []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for index := 0; index < len(set); index++ {
		_, known := seen[set[index]]
		if !known {
			seen[set[index]] = true
			result = append(result, set[index])
		}
	}
	return result
}

// complement returns, in ascending order, every ASCII character that is not in
// set, followed by each non-ASCII character of data that is not in set.
func complement(set []string, data string) []string {
	member := map[string]bool{}
	for index := 0; index < len(set); index++ {
		member[set[index]] = true
	}
	result := []string{}
	for code := 0; code < 128; code++ {
		_, known := member[cli.Char(code)]
		if !known {
			result = append(result, cli.Char(code))
		}
	}
	if cli.IsASCII(data) {
		return result
	}
	rest := data
	for code := 0; code < 128; code++ {
		rest = cli.ReplaceAll(rest, cli.Char(code), "")
	}
	for rest != "" {
		width := 2
		if rest[0] >= 240 {
			width = 4
		} else if rest[0] >= 224 {
			width = 3
		}
		character := rest[:width]
		rest = cli.ReplaceAll(rest, character, "")
		_, known := member[character]
		if !known {
			result = append(result, character)
		}
	}
	return result
}

func main() {
	flags, values, operands := cli.ParseOptions("tr", usage, os.Args[1:], "cCdst", "complement=c delete=d squeeze-repeats=s truncate-set1=t", cli.ModePlain, 1)
	complemented := false
	deleting := false
	squeezing := false
	truncating := false
	for index := 0; index < len(flags); index++ {
		flag := flags[index] + values[index]
		if flag == "c" || flag == "C" {
			complemented = true
		} else if flag == "d" {
			deleting = true
		} else if flag == "s" {
			squeezing = true
		} else {
			truncating = true
		}
	}
	if len(operands) == 0 {
		cli.FailUsage("tr", "missing operand", 1)
	}
	translating := !deleting && len(operands) >= 2
	if len(operands) == 1 && !deleting && !squeezing {
		cli.FailUsage("tr", "missing operand after '"+operands[0]+"'\nTwo strings must be given when translating.", 1)
	}
	if len(operands) == 1 && deleting && squeezing {
		cli.FailUsage("tr", "missing operand after '"+operands[0]+"'\nTwo strings must be given when both deleting and squeezing repeats.", 1)
	}
	limit := 2
	if deleting && !squeezing {
		limit = 1
	}
	if len(operands) > limit {
		cli.FailUsage("tr", "extra operand '"+operands[limit]+"'", 1)
	}

	data := cli.ReadStdin()
	first := expandSet(operands[0])
	if complemented {
		first = complement(first, data)
	}
	second := []string{}
	if len(operands) == 2 {
		second = expandSet(operands[1])
		if setRepeatAt >= 0 {
			padded := []string{}
			for index := 0; index < setRepeatAt; index++ {
				padded = append(padded, second[index])
			}
			for len(padded)+len(second)-setRepeatAt < len(first) {
				padded = append(padded, setRepeatChar)
			}
			for index := setRepeatAt; index < len(second); index++ {
				padded = append(padded, second[index])
			}
			second = padded
		}
	}

	if deleting {
		members := distinct(first)
		for index := 0; index < len(members); index++ {
			data = cli.ReplaceAll(data, members[index], "")
		}
	}
	if translating {
		if len(second) == 0 {
			cli.Fail("tr", "when not truncating set1, string2 must be non-empty", 1)
		}
		if truncating && len(first) > len(second) {
			first = first[:len(second)]
		}
		// The last mapping of a repeated SET1 character wins, as in GNU tr.
		target := map[string]string{}
		sources := []string{}
		for index := 0; index < len(first); index++ {
			to := second[len(second)-1]
			if index < len(second) {
				to = second[index]
			}
			_, known := target[first[index]]
			if !known {
				sources = append(sources, first[index])
			}
			target[first[index]] = to
		}
		// One pass is enough unless some target is itself translated to
		// something else; then go through placeholders.
		direct := true
		for index := 0; index < len(sources); index++ {
			to := target[sources[index]]
			onward, mapped := target[to]
			if mapped && onward != to {
				direct = false
			}
		}
		if direct {
			for index := 0; index < len(sources); index++ {
				data = cli.ReplaceAll(data, sources[index], target[sources[index]])
			}
		} else {
			moved := []string{}
			for index := 0; index < len(sources); index++ {
				from := sources[index]
				if !strings.Contains(data, from) {
					continue
				}
				slot := len(moved)
				if slot >= 256 {
					cli.Fail("tr", "too many overlapping translations", 1)
				}
				holder := placeholders[slot*3 : slot*3+3]
				if strings.Contains(data, holder) {
					cli.Fail("tr", "input contains private-use characters needed for an overlapping translation", 1)
				}
				data = cli.ReplaceAll(data, from, holder)
				moved = append(moved, from)
			}
			for slot := 0; slot < len(moved); slot++ {
				data = cli.ReplaceAll(data, placeholders[slot*3:slot*3+3], target[moved[slot]])
			}
		}
	}
	if squeezing {
		set := first
		if len(operands) == 2 {
			set = second
		}
		members := distinct(set)
		for index := 0; index < len(members); index++ {
			data = cli.SqueezeRuns(data, members[index])
		}
	}
	fmt.Print(data)
	os.Exit(0)
}
