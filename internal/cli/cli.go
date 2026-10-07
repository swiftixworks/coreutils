// Package cli is the code every Swiftix coreutils command shares: the option
// parser, operand and standard-input readers, diagnostics, and string helpers.
//
// A native `strings` call costs a few VM instructions regardless of input size,
// so commands prefer whole-buffer Split/Join/Count/Index passes to guest loops
// where both are equally clear.
//
// Results go to standard output with `fmt.Print` (text) or `os.Stdout.Write`
// (exact bytes); neither is buffered here, and both wait while a pipe is full.
// Diagnostics and usage errors go to standard error. `--help` is a result.
package cli

import "fmt"
import "os"
import "strconv"
import "strings"
import "swiftix/userland"

// asciiTable holds the 128 ASCII characters in code order; latinTable holds
// U+0080...U+00FF, two UTF-8 bytes each.
const asciiTable = "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~\x7f"
const latinTable = "\u0080\u0081\u0082\u0083\u0084\u0085\u0086\u0087\u0088\u0089\u008a\u008b\u008c\u008d\u008e\u008f\u0090\u0091\u0092\u0093\u0094\u0095\u0096\u0097\u0098\u0099\u009a\u009b\u009c\u009d\u009e\u009f\u00a0\u00a1\u00a2\u00a3\u00a4\u00a5\u00a6\u00a7\u00a8\u00a9\u00aa\u00ab\u00ac\u00ad\u00ae\u00af\u00b0\u00b1\u00b2\u00b3\u00b4\u00b5\u00b6\u00b7\u00b8\u00b9\u00ba\u00bb\u00bc\u00bd\u00be\u00bf\u00c0\u00c1\u00c2\u00c3\u00c4\u00c5\u00c6\u00c7\u00c8\u00c9\u00ca\u00cb\u00cc\u00cd\u00ce\u00cf\u00d0\u00d1\u00d2\u00d3\u00d4\u00d5\u00d6\u00d7\u00d8\u00d9\u00da\u00db\u00dc\u00dd\u00de\u00df\u00e0\u00e1\u00e2\u00e3\u00e4\u00e5\u00e6\u00e7\u00e8\u00e9\u00ea\u00eb\u00ec\u00ed\u00ee\u00ef\u00f0\u00f1\u00f2\u00f3\u00f4\u00f5\u00f6\u00f7\u00f8\u00f9\u00fa\u00fb\u00fc\u00fd\u00fe\u00ff"

const lowerLetters = "abcdefghijklmnopqrstuvwxyz"
const upperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// ModePlain makes ParseOptions treat "-NUM" like any other option cluster.
const ModePlain = 0

// ModeCount makes ParseOptions turn "-NUM" into option "#" with value NUM.
const ModeCount = 1

// ModeNumbers makes ParseOptions keep "-NUM" as an operand.
const ModeNumbers = 2

// Char returns the character with the given code: ASCII below 128 and the
// Latin-1 code point up to 255. Other codes yield the empty string.
func Char(code int) string {
	if code >= 0 && code < 128 {
		return asciiTable[code : code+1]
	}
	if code >= 128 && code < 256 {
		offset := (code - 128) * 2
		return latinTable[offset : offset+2]
	}
	return ""
}

// ReplaceAll returns text with every non-overlapping occurrence of from
// replaced by to. It runs as two native calls.
func ReplaceAll(text string, from string, to string) string {
	if from == "" || from == to {
		return text
	}
	return strings.Join(strings.Split(text, from), to)
}

// SqueezeRuns collapses every run of unit in text to a single unit.
func SqueezeRuns(text string, unit string) string {
	double := unit + unit
	for strings.Contains(text, double) {
		text = strings.Join(strings.Split(text, double), unit)
	}
	return text
}

// UpperASCII folds ASCII lower-case letters to upper case.
func UpperASCII(text string) string {
	for index := 0; index < 26; index++ {
		text = ReplaceAll(text, lowerLetters[index:index+1], upperLetters[index:index+1])
	}
	return text
}

// LowerASCII folds ASCII upper-case letters to lower case.
func LowerASCII(text string) string {
	for index := 0; index < 26; index++ {
		text = ReplaceAll(text, upperLetters[index:index+1], lowerLetters[index:index+1])
	}
	return text
}

// IsASCII reports whether text contains only single-byte characters.
func IsASCII(text string) bool {
	return strings.Count(text, "") == len(text)+1
}

// CharCount returns the number of characters in text.
func CharCount(text string) int {
	return strings.Count(text, "") - 1
}

// DigitCount returns the number of characters needed to print value.
func DigitCount(value int) int {
	count := 1
	if value < 0 {
		count = 2
		value = 0 - value
	}
	for value >= 10 {
		value = value / 10
		count++
	}
	return count
}

// PadLeft right-aligns text in a field of the given width.
func PadLeft(text string, width int) string {
	missing := width - CharCount(text)
	if missing <= 0 {
		return text
	}
	return strings.Repeat(" ", missing) + text
}

// IsDigits reports whether text is a non-empty run of decimal digits.
func IsDigits(text string) bool {
	if text == "" {
		return false
	}
	for index := 0; index < len(text); index++ {
		if text[index] < 48 || text[index] > 57 {
			return false
		}
	}
	return true
}

// ParseCount parses a non-negative decimal count. Values too large for an int
// saturate, which every caller treats as "everything".
func ParseCount(text string) (int, bool) {
	if !IsDigits(text) {
		return 0, false
	}
	if len(text) > 15 {
		return 1000000000000000, true
	}
	value, _ := strconv.Atoi(text)
	return value, true
}

// ParseSize parses a count with an optional multiplier suffix: b (512), k or K
// or KiB (1024), kB or KB (1000), m or M or MiB (1048576), MB (1000000).
func ParseSize(text string) (int, bool) {
	end := 0
	for end < len(text) && text[end] >= 48 && text[end] <= 57 {
		end++
	}
	value, ok := ParseCount(text[:end])
	if !ok {
		return 0, false
	}
	suffix := text[end:]
	scale := 1
	if suffix == "" {
		scale = 1
	} else if suffix == "b" {
		scale = 512
	} else if suffix == "k" || suffix == "K" || suffix == "KiB" {
		scale = 1024
	} else if suffix == "kB" || suffix == "KB" {
		scale = 1000
	} else if suffix == "m" || suffix == "M" || suffix == "MiB" {
		scale = 1048576
	} else if suffix == "MB" {
		scale = 1000000
	} else {
		return 0, false
	}
	return value * scale, true
}

// Fail prints "name: message" on standard error and exits with status.
func Fail(name string, message string, status int) {
	fmt.Fprint(os.Stderr, name+": "+message+"\n")
	os.Exit(status)
}

// FailUsage prints "name: message" on standard error, points at --help, and
// exits with status.
func FailUsage(name string, message string, status int) {
	fmt.Fprint(os.Stderr, name+": "+message+"\nTry '"+name+" --help' for more information.\n")
	os.Exit(status)
}

// longOption returns the short letter that longs maps name to, or "".
func longOption(longs string, name string) string {
	entries := strings.Split(longs, " ")
	for index := 0; index < len(entries); index++ {
		entry := entries[index]
		split := strings.Index(entry, "=")
		if split > 0 && entry[:split] == name {
			return entry[split+1:]
		}
	}
	return ""
}

// takesValue reports whether spec marks letter as taking a value.
func takesValue(spec string, letter string) bool {
	position := strings.Index(spec, letter)
	return position >= 0 && position+1 < len(spec) && spec[position+1] == 58
}

// ParseOptions splits args into options and operands the way GNU getopt does.
//
// spec lists the short options; a letter followed by ':' takes a value. longs
// maps long names to short letters as space-separated "name=letter" entries.
// Options may be combined ("-rn"), carry attached values ("-n5", "--lines=5"),
// and appear after operands; "--" ends option parsing and "-" is an operand.
// "--help" prints usage on standard output and exits 0. An unknown option or a
// missing value prints a diagnostic on standard error and exits with failStatus.
//
// mode selects extra syntax: ModeCount turns "-NUM" into option "#" with value
// NUM (head, tail), and ModeNumbers keeps "-NUM" as an operand (seq).
//
// The results are parallel slices of option letters and their values (empty
// for flags), in command-line order, followed by the operands.
func ParseOptions(name string, usage string, args []string, spec string, longs string, mode int, failStatus int) ([]string, []string, []string) {
	flags := []string{}
	values := []string{}
	operands := []string{}
	done := false
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if done || len(arg) < 2 || arg[0] != 45 {
			operands = append(operands, arg)
			continue
		}
		if arg == "--" {
			done = true
			continue
		}
		numeric := (arg[1] >= 48 && arg[1] <= 57) || arg[1] == 46
		if mode == ModeNumbers && numeric {
			operands = append(operands, arg)
			continue
		}
		if mode == ModeCount && numeric {
			flags = append(flags, "#")
			values = append(values, arg[1:])
			continue
		}
		if arg[1] == 45 {
			body := arg[2:]
			value := ""
			hasValue := false
			split := strings.Index(body, "=")
			if split >= 0 {
				value = body[split+1:]
				body = body[:split]
				hasValue = true
			}
			if body == "help" {
				fmt.Print(usage)
				os.Exit(0)
			}
			letter := longOption(longs, body)
			if letter == "" {
				FailUsage(name, "unrecognized option '--"+body+"'", failStatus)
			}
			if takesValue(spec, letter) {
				if !hasValue {
					index++
					if index >= len(args) {
						FailUsage(name, "option '--"+body+"' requires an argument", failStatus)
					}
					value = args[index]
				}
			} else if hasValue {
				FailUsage(name, "option '--"+body+"' doesn't allow an argument", failStatus)
			}
			flags = append(flags, letter)
			values = append(values, value)
			continue
		}
		for position := 1; position < len(arg); position++ {
			if arg[position] >= 128 || arg[position] == 58 || arg[position] == 35 {
				FailUsage(name, "invalid option -- '"+arg[position:]+"'", failStatus)
			}
			letter := arg[position : position+1]
			if strings.Index(spec, letter) < 0 {
				FailUsage(name, "invalid option -- '"+letter+"'", failStatus)
			}
			if takesValue(spec, letter) {
				value := ""
				if position+1 < len(arg) {
					value = arg[position+1:]
				} else {
					index++
					if index >= len(args) {
						FailUsage(name, "option requires an argument -- '"+letter+"'", failStatus)
					}
					value = args[index]
				}
				flags = append(flags, letter)
				values = append(values, value)
				break
			}
			flags = append(flags, letter)
			values = append(values, "")
		}
	}
	return flags, values, operands
}

// ReadStdin reads standard input to end of file as text. Each read waits for
// the producer, whether that is a pipe or someone typing at a terminal.
func ReadStdin() string {
	parts := []string{}
	for {
		chunk, status := userland.ReadStdin()
		if status != 0 {
			break
		}
		parts = append(parts, chunk)
	}
	return strings.Join(parts, "")
}

// readStdinBytes reads standard input as exact bytes, in the chunks the reads
// returned, and stops after limit bytes; a negative limit reads to end of file.
// It returns the chunks and their total length.
func readStdinBytes(limit int) ([][]byte, int) {
	chunks := [][]byte{}
	total := 0
	for limit < 0 || total < limit {
		size := 65536
		if limit >= 0 && limit-total < size {
			size = limit - total
		}
		buffer := make([]byte, size)
		count, err := os.Stdin.Read(buffer)
		if count > 0 {
			chunks = append(chunks, buffer[:count])
			total = total + count
		}
		if err != nil {
			break
		}
	}
	return chunks, total
}

// ReadBytes reads one operand as exact bytes: "-" is standard input (up to
// limit bytes, or all of it when limit is negative), anything else a path that
// is read whole. A file that cannot be read is reported on standard error and
// yields ok false.
func ReadBytes(name string, path string, limit int) ([][]byte, int, bool) {
	if path == "-" {
		chunks, total := readStdinBytes(limit)
		return chunks, total, true
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprint(os.Stderr, name+": "+path+": No such file\n")
		return [][]byte{}, 0, false
	}
	return [][]byte{data}, len(data), true
}

// WriteBytes writes bytes from up to (not including) to of the concatenated
// chunks to standard output, unchanged.
func WriteBytes(chunks [][]byte, from int, to int) {
	offset := 0
	for index := 0; index < len(chunks); index++ {
		chunk := chunks[index]
		end := offset + len(chunk)
		if end > from && offset < to {
			low := 0
			if from > offset {
				low = from - offset
			}
			high := len(chunk)
			if to < end {
				high = to - offset
			}
			os.Stdout.Write(chunk[low:high])
		}
		offset = end
	}
}

// ReadFile reads one operand as text: "-" is standard input, anything else a
// path. A missing file is reported on standard error by the runtime and yields
// status 1.
func ReadFile(name string, path string) (string, int) {
	if path == "-" {
		return ReadStdin(), 0
	}
	return userland.ReadInput(name, []string{path})
}

// ReadAll concatenates the operands, or reads standard input when there are
// none. The status is 1 when any file could not be read.
func ReadAll(name string, operands []string) (string, int) {
	if len(operands) == 0 {
		return ReadStdin(), 0
	}
	if len(operands) == 1 {
		return ReadFile(name, operands[0])
	}
	plain := true
	for index := 0; index < len(operands); index++ {
		if operands[index] == "-" {
			plain = false
		}
	}
	if plain {
		return userland.ReadInput(name, operands)
	}
	parts := []string{}
	result := 0
	for index := 0; index < len(operands); index++ {
		data, status := ReadFile(name, operands[index])
		parts = append(parts, data)
		if status != 0 {
			result = status
		}
	}
	return strings.Join(parts, ""), result
}

// SplitLines splits data into lines without their terminators. A final line
// that lacks a newline is still a line; empty data has no lines.
func SplitLines(data string) []string {
	if data == "" {
		return []string{}
	}
	if strings.HasSuffix(data, "\n") {
		data = data[:len(data)-1]
	}
	return strings.Split(data, "\n")
}

// LineCount returns the number of lines in data, counting an unterminated
// final line.
func LineCount(data string) int {
	count := strings.Count(data, "\n")
	if data != "" && !strings.HasSuffix(data, "\n") {
		count++
	}
	return count
}

// LineOffset returns the byte offset just past the first count lines of data,
// or len(data) when data has fewer lines.
func LineOffset(data string, count int) int {
	if count <= 0 {
		return 0
	}
	total := strings.Count(data, "\n")
	if count > total {
		return len(data)
	}
	if count == total {
		return strings.LastIndex(data, "\n") + 1
	}
	if count <= 24 {
		offset := 0
		rest := data
		for index := 0; index < count; index++ {
			next := strings.Index(rest, "\n")
			offset = offset + next + 1
			rest = rest[next+1:]
		}
		return offset
	}
	lines := strings.Split(data, "\n")
	return len(strings.Join(lines[:count], "\n")) + 1
}
