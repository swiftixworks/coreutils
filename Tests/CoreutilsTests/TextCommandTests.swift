import Testing

/// Every documented option of the line-oriented commands, executed from the
/// packaged images with exact standard output, standard error and exit status.
@Suite("text commands")
struct TextCommandTests {
    @Test("true and false only set the exit status")
    func trueAndFalse() throws {
        try withGuest { guest in
            expect(guest, "true", "")
            expect(guest, "false", "", status: 1)
            expect(guest, "true --help", "")
            expect(guest, "false ignored operands", "", status: 1)
        }
    }

    @Test("echo handles -n, -e, -E and backslash escapes")
    func echo() throws {
        try withGuest { guest in
            expect(guest, "echo hello world", "hello world\n")
            expect(guest, "echo", "\n")
            expect(guest, "echo -n x", "x")
            expect(guest, #"echo -e 'a\tb'"#, "a\tb\n")
            expect(guest, #"echo -E 'a\tb'"#, "a\\tb\n")
            expect(guest, #"echo 'a\tb'"#, "a\\tb\n")
            expect(guest, #"echo -ne 'a\nb\n'"#, "a\nb\n")
            expect(guest, #"echo -n -e 'q\n'"#, "q\n")
            expect(guest, #"echo -e -E 'q\n'"#, "q\\n\n")
            expect(guest, #"echo -e 'x\\y'"#, "x\\y\n")
            expect(guest, #"echo -e 'a\ab\bc\rd\fe\vf\eg'"#, "a\u{07}b\u{08}c\rd\u{0C}e\u{0B}f\u{1B}g\n")
            expect(guest, #"echo -e '\0101\x41\x4a\0'"#, "AAJ\0\n")
            expect(guest, #"echo -e 'one\ctwo' three"#, "one")
            expect(guest, #"echo -e 'keep \q \'"#, "keep \\q \\\n")
            expect(guest, "echo -x", "-x\n")
            expect(guest, "echo - a", "- a\n")
            expect(guest, "echo -- -n", "-- -n\n")
            expect(guest, "echo a -n", "a -n\n")
            expect(guest, "echo -nn x", "x")
        }
    }

    @Test("cat concatenates, numbers, squeezes and reveals characters")
    func cat() throws {
        try withGuest(files: Fixture.files.merging( ["/control": "a\u{01}b\tc\u{7F}\n"]) { $1 }) { guest in
            expect(guest, "cat /left /right", "a\nc\ne\nb\nc\nd\n")
            expect(guest, "cat /blanks", "one\n\n\n\ntwo\tx\n\nthree")
            expect(
                guest, "cat -n /left /right",
                "     1\ta\n     2\tc\n     3\te\n     4\tb\n     5\tc\n     6\td\n")
            expect(guest, "cat --number /left", "     1\ta\n     2\tc\n     3\te\n")
            expect(
                guest, "cat -b /blanks",
                "     1\tone\n\n\n\n     2\ttwo\tx\n\n     3\tthree")
            expect(guest, "cat -nb /left", "     1\ta\n     2\tc\n     3\te\n")
            expect(guest, "cat -s /blanks", "one\n\ntwo\tx\n\nthree")
            expect(
                guest, "cat -sn /blanks",
                "     1\tone\n     2\t\n     3\ttwo\tx\n     4\t\n     5\tthree")
            expect(guest, "cat -E /left", "a$\nc$\ne$\n")
            expect(guest, "cat -T /blanks", "one\n\n\n\ntwo^Ix\n\nthree")
            expect(guest, "cat -A /blanks", "one$\n$\n$\n$\ntwo^Ix$\n$\nthree")
            expect(guest, "cat -v /control", "a^Ab\tc^?\n")
            expect(guest, "cat -e /control", "a^Ab\tc^?$\n")
            expect(guest, "cat -t /control", "a^Ab^Ic^?\n")
            expect(guest, "cat -nE /left", "     1\ta$\n     2\tc$\n     3\te$\n")
            expect(guest, "cat -u /left", "a\nc\ne\n")
            expect(guest, "cat -- /left", "a\nc\ne\n")
            expect(guest, "cat /left -n", "     1\ta\n     2\tc\n     3\te\n")
            expect(guest, "cat < /left", "a\nc\ne\n")
            expect(guest, "cat -n < /left", "     1\ta\n     2\tc\n     3\te\n")
            expect(guest, "echo piped | cat - /left", "piped\na\nc\ne\n")
            expect(guest, "echo piped | cat /left -", "a\nc\ne\npiped\n")
            expect(guest, "echo piped | cat", "piped\n")
            expect(
                guest, "cat /left /missing", "a\nc\ne\n", status: 1,
                errors: "cat: /missing: No such file\n")
            expect(
                guest, "cat -x /left",
                "", status: 1,
                errors: "cat: invalid option -- 'x'\nTry 'cat --help' for more information.\n")
            expect(
                guest, "cat --bogus",
                "", status: 1,
                errors: "cat: unrecognized option '--bogus'\nTry 'cat --help' for more information.\n")
        }
    }

    @Test("head selects leading lines and bytes")
    func head() throws {
        try withGuest(files: Fixture.files) { guest in
            let lines = (1...12).map { "line \($0)\n" }
            expect(guest, "head /ten", lines[..<10].joined())
            expect(guest, "head -n 3 /ten", lines[..<3].joined())
            expect(guest, "head -n3 /ten", lines[..<3].joined())
            expect(guest, "head -3 /ten", lines[..<3].joined())
            expect(guest, "head --lines=3 /ten", lines[..<3].joined())
            expect(guest, "head --lines 3 /ten", lines[..<3].joined())
            expect(guest, "head -n -10 /ten", lines[..<2].joined())
            expect(guest, "head -n 0 /ten", "")
            expect(guest, "head -n 100 /ten", lines.joined())
            expect(guest, "head -n 2 /blanks", "one\n\n")
            expect(guest, "head -n 30 /blanks", "one\n\n\n\ntwo\tx\n\nthree")
            expect(guest, "head -c 6 /ten", "line 1")
            expect(guest, "head -c6 /ten", "line 1")
            expect(guest, "head --bytes=6 /ten", "line 1")
            expect(guest, "head -c -4 /left", "a\n")
            expect(guest, "head -c 1K /left", "a\nc\ne\n")
            expect(guest, "head -c 3 /utf8", "hé")
            expect(guest, "head -n 1 /left /right", "==> /left <==\na\n\n==> /right <==\nb\n")
            expect(guest, "head -q -n 1 /left /right", "a\nb\n")
            expect(guest, "head -v -n 1 /left", "==> /left <==\na\n")
            expect(guest, "cat /ten | head -n 2", lines[..<2].joined())
            expect(guest, "head -n 2 - < /ten", lines[..<2].joined())
            expect(guest, "cat /ten | head -c 4", "line")
            expect(guest, "cat /ten | head -n -11", lines[..<1].joined())
            expect(
                guest, "head -n 1 /missing /left", "==> /left <==\na\n", status: 1,
                errors: "head: /missing: No such file\n")
            expect(guest, "head -n x /left", "", status: 1,
                errors: "head: invalid number of lines: 'x'\n")
            expect(guest, "head -c x /left", "", status: 1,
                errors: "head: invalid number of bytes: 'x'\n")
            expect(
                guest, "head -n",
                "", status: 1,
                errors: "head: option requires an argument -- 'n'\nTry 'head --help' for more information.\n")
        }
    }

    @Test("tail selects trailing lines and bytes")
    func tail() throws {
        try withGuest(files: Fixture.files) { guest in
            let lines = (1...12).map { "line \($0)\n" }
            expect(guest, "tail /ten", lines[2...].joined())
            expect(guest, "tail -n 2 /ten", lines[10...].joined())
            expect(guest, "tail -n2 /ten", lines[10...].joined())
            expect(guest, "tail -2 /ten", lines[10...].joined())
            expect(guest, "tail --lines=2 /ten", lines[10...].joined())
            expect(guest, "tail -n -2 /ten", lines[10...].joined())
            expect(guest, "tail -n +11 /ten", lines[10...].joined())
            expect(guest, "tail -n +1 /left", "a\nc\ne\n")
            expect(guest, "tail -n 0 /ten", "")
            expect(guest, "tail -n 100 /left", "a\nc\ne\n")
            expect(guest, "tail -n 2 /blanks", "\nthree")
            expect(guest, "tail -n 30 /ten", lines.joined())
            expect(guest, "tail -c 3 /left", "\ne\n")
            expect(guest, "tail -c3 /left", "\ne\n")
            expect(guest, "tail --bytes=3 /left", "\ne\n")
            expect(guest, "tail -c +5 /left", "e\n")
            expect(guest, "tail -c 100 /left", "a\nc\ne\n")
            expect(guest, "tail -n 1 /left /right", "==> /left <==\ne\n\n==> /right <==\nd\n")
            expect(guest, "tail -q -n 1 /left /right", "e\nd\n")
            expect(guest, "tail -v -n 1 /left", "==> /left <==\ne\n")
            expect(guest, "cat /ten | tail -n 1", "line 12\n")
            expect(guest, "tail -n 1 - < /ten", "line 12\n")
            expect(guest, "cat /ten | tail -c 3", "12\n")
            expect(
                guest, "tail -n 1 /missing /left", "==> /left <==\ne\n", status: 1,
                errors: "tail: /missing: No such file\n")
            expect(guest, "tail -n x /left", "", status: 1,
                errors: "tail: invalid number of lines: 'x'\n")
            expect(
                guest, "tail -z /left",
                "", status: 1,
                errors: "tail: invalid option -- 'z'\nTry 'tail --help' for more information.\n")
        }
    }

    @Test("wc counts lines, words, characters, bytes and line width")
    func wc() throws {
        try withGuest(files: Fixture.files) { guest in
            expect(guest, "wc /fruits", " 4  4 26 /fruits\n")
            expect(guest, "wc -l /fruits", "4 /fruits\n")
            expect(guest, "wc -w /fruits", "4 /fruits\n")
            expect(guest, "wc -c /fruits", "26 /fruits\n")
            expect(guest, "wc --lines /fruits", "4 /fruits\n")
            expect(guest, "wc --words --bytes /fruits", " 4 26 /fruits\n")
            expect(guest, "wc -c /utf8", "24 /utf8\n")
            expect(guest, "wc -m /utf8", "22 /utf8\n")
            expect(guest, "wc -L /utf8", "16 /utf8\n")
            expect(guest, "wc -L /fruits", "6 /fruits\n")
            expect(guest, "wc -w /blanks", "4 /blanks\n")
            expect(guest, "wc -l /blanks", "6 /blanks\n")
            expect(guest, "wc -lw /fruits", "4 4 /fruits\n")
            expect(guest, "wc -lwcmL /utf8", " 2  4 22 24 16 /utf8\n")
            expect(guest, "wc -l < /fruits", "4\n")
            expect(guest, "cat /fruits | wc -l", "4\n")
            expect(guest, "cat /fruits | wc -w", "4\n")
            expect(guest, "cat /fruits | wc -c", "26\n")
            expect(guest, "cat /fruits | wc", "      4       4      26\n")
            expect(guest, "cat /fruits | wc -l -", "4 -\n")
            expect(guest, "wc /fruits /left", " 4  4 26 /fruits\n 3  3  6 /left\n 7  7 32 total\n")
            expect(guest, "wc -l /fruits /left", "4 /fruits\n3 /left\n7 total\n")
            expect(
                guest, "wc -l /fruits /missing", "4 /fruits\n4 total\n", status: 1,
                errors: "wc: /missing: No such file\n")
            expect(guest, "echo -n | wc", "      0       0       0\n")
            expect(guest, "echo -n 'no newline' | wc -lw", "      0       2\n")
            expect(
                guest, "wc -q /fruits",
                "", status: 1,
                errors: "wc: invalid option -- 'q'\nTry 'wc --help' for more information.\n")
            // The original defect: wc treated "-l" as a file name.
            expect(guest, "ls /usr/bin | wc -w", "17\n")
            expect(guest, "seq 17 | wc -l", "17\n")
        }
    }

    @Test("tac and rev reverse lines and characters")
    func tacAndRev() throws {
        try withGuest(files: Fixture.files) { guest in
            expect(guest, "tac /left", "e\nc\na\n")
            expect(guest, "tac /left /right", "e\nc\na\nd\nc\nb\n")
            expect(guest, "tac /blanks", "three\ntwo\tx\n\n\n\none\n")
            expect(guest, "cat /left | tac", "e\nc\na\n")
            expect(guest, "tac - < /left", "e\nc\na\n")
            expect(guest, "echo a,b,c | tac -s ,", "c\nb,a,")
            expect(guest, "echo -n a,b,c, | tac --separator=,", "c,b,a,")
            expect(guest, "tac /missing", "", status: 1, errors: "tac: /missing: No such file\n")
            expect(guest, "rev /fruits", "ananab\nelppa\nyrrehc\nelppa\n")
            expect(guest, "rev /utf8", "dlröw olléh\nrab  oof\t\n")
            expect(guest, "echo abc | rev", "cba\n")
            expect(guest, "echo -n abc | rev", "cba")
            expect(guest, "rev /left /right", "a\nc\ne\nb\nc\nd\n")
            expect(guest, "echo abc | rev - /left", "cba\na\nc\ne\n")
        }
    }

    @Test("nl numbers lines in every style and format")
    func nl() throws {
        try withGuest(files: Fixture.files) { guest in
            expect(guest, "nl /left", "     1\ta\n     2\tc\n     3\te\n")
            expect(
                guest, "nl /blanks",
                "     1\tone\n       \n       \n       \n     2\ttwo\tx\n       \n     3\tthree")
            expect(
                guest, "nl -b a /blanks",
                "     1\tone\n     2\t\n     3\t\n     4\t\n     5\ttwo\tx\n     6\t\n     7\tthree")
            expect(guest, "nl -ba -w3 -s: /left", "  1:a\n  2:c\n  3:e\n")
            expect(guest, "nl -b t /left", "     1\ta\n     2\tc\n     3\te\n")
            expect(guest, "nl -b n /left", "       a\n       c\n       e\n")
            expect(guest, "nl -w 2 /left", " 1\ta\n 2\tc\n 3\te\n")
            expect(guest, "nl -s ') ' /left", "     1) a\n     2) c\n     3) e\n")
            expect(guest, "nl -v 10 /left", "    10\ta\n    11\tc\n    12\te\n")
            expect(guest, "nl -v10 -i5 /left", "    10\ta\n    15\tc\n    20\te\n")
            expect(guest, "nl -n ln /left", "1     \ta\n2     \tc\n3     \te\n")
            expect(guest, "nl -n rn /left", "     1\ta\n     2\tc\n     3\te\n")
            expect(guest, "nl -n rz -w 4 /left", "0001\ta\n0002\tc\n0003\te\n")
            expect(guest, "nl -v -1 -n rz -w 3 /left", "-01\ta\n000\tc\n001\te\n")
            expect(guest, "cat /left | nl", "     1\ta\n     2\tc\n     3\te\n")
            expect(
                guest, "nl --body-numbering=a --number-width=2 --number-separator=. /left",
                " 1.a\n 2.c\n 3.e\n")
            expect(guest, "nl -b x /left", "", status: 1,
                errors: "nl: invalid body numbering style: 'x'\n")
            expect(guest, "nl -n zz /left", "", status: 1,
                errors: "nl: invalid line numbering format: 'zz'\n")
        }
    }

    @Test("fold wraps long lines by characters, bytes and at spaces")
    func fold() throws {
        try withGuest(files: Fixture.files) { guest in
            expect(guest, "echo abcdefghij | fold -w 4", "abcd\nefgh\nij\n")
            expect(guest, "echo abcdefghij | fold -w4", "abcd\nefgh\nij\n")
            expect(guest, "echo abcdefghij | fold -4", "abcd\nefgh\nij\n")
            expect(guest, "echo abcdefghij | fold --width=5", "abcde\nfghij\n")
            expect(guest, "echo 'aa bb cc dd' | fold -s -w 5", "aa \nbb \ncc dd\n")
            expect(guest, "echo 'aa bb cc dd' | fold -sw 5", "aa \nbb \ncc dd\n")
            expect(guest, "echo abcdefgh | fold -s -w 3", "abc\ndef\ngh\n")
            expect(guest, "fold -w 3 /fruits", "ban\nana\napp\nle\nche\nrry\napp\nle\n")
            expect(guest, "fold /fruits", "banana\napple\ncherry\napple\n")
            expect(guest, "fold -w 5 /blanks", "one\n\n\n\ntwo\tx\n\nthree")
            expect(guest, "echo héllo | fold -w 2", "hé\nll\no\n")
            expect(guest, "echo héllo | fold -b -w 2", "h\né\nll\no\n")
            expect(guest, "echo -n abcdef | fold -w 4", "abcd\nef")
            expect(
                guest, "echo \(String(repeating: "x", count: 100)) | fold",
                String(repeating: "x", count: 80) + "\n" + String(repeating: "x", count: 20) + "\n")
            expect(guest, "fold -w 0 /left", "", status: 1,
                errors: "fold: invalid number of columns: '0'\n")
        }
    }

    @Test("paste joins lines in parallel and serially")
    func paste() throws {
        try withGuest(files: Fixture.files) { guest in
            expect(guest, "paste /left /right", "a\tb\nc\tc\ne\td\n")
            expect(guest, "paste -d , /left /right", "a,b\nc,c\ne,d\n")
            expect(guest, "paste -d, /left /right /left", "a,b,a\nc,c,c\ne,d,e\n")
            expect(guest, "paste -d ,: /left /right /left", "a,b:a\nc,c:c\ne,d:e\n")
            expect(guest, "paste --delimiters=- /left /right", "a-b\nc-c\ne-d\n")
            expect(guest, "paste -s /left /right", "a\tc\te\nb\tc\td\n")
            expect(guest, "paste -s -d , /left", "a,c,e\n")
            expect(guest, "paste -sd ,: /left", "a,c:e\n")
            expect(guest, "paste --serial /left", "a\tc\te\n")
            expect(guest, #"paste -d '\n' /left /right"#, "a\nb\nc\nc\ne\nd\n")
            expect(guest, #"paste -d '\0' /left /right"#, "ab\ncc\ned\n")
            expect(guest, "paste /left /fruits", "a\tbanana\nc\tapple\ne\tcherry\n\tapple\n")
            expect(guest, "seq 6 | paste - - -", "1\t2\t3\n4\t5\t6\n")
            expect(guest, "seq 5 | paste - -", "1\t2\n3\t4\n5\t\n")
            expect(guest, "cat /left | paste - /right", "a\tb\nc\tc\ne\td\n")
            expect(guest, "cat /left | paste", "a\nc\ne\n")
            expect(guest, "seq 4 | paste -s -d + -", "1+2+3+4\n")
        }
    }

    @Test("comm compares two sorted files in three columns")
    func comm() throws {
        try withGuest(files: Fixture.files) { guest in
            expect(guest, "comm /left /right", "a\n\tb\n\t\tc\n\td\ne\n")
            expect(guest, "comm -1 /left /right", "b\n\tc\nd\n")
            expect(guest, "comm -2 /left /right", "a\n\tc\ne\n")
            expect(guest, "comm -3 /left /right", "a\n\tb\n\td\ne\n")
            expect(guest, "comm -12 /left /right", "c\n")
            expect(guest, "comm -1 -2 /left /right", "c\n")
            expect(guest, "comm -13 /left /right", "b\nd\n")
            expect(guest, "comm -23 /left /right", "a\ne\n")
            expect(guest, "comm -123 /left /right", "")
            expect(guest, "comm -23 /left - < /right", "a\ne\n")
            expect(guest, "cat /left | comm -13 - /right", "b\nd\n")
            expect(guest, "comm --output-delimiter=: /left /right", "a\n:b\n::c\n:d\ne\n")
            expect(guest, "comm --total /left /right", "a\n\tb\n\t\tc\n\td\ne\n2\t2\t1\ttotal\n")
            expect(
                guest, "comm /left",
                "", status: 1,
                errors: "comm: missing operand after '/left'\nTry 'comm --help' for more information.\n")
            expect(
                guest, "comm /left /missing", "", status: 1,
                errors: "comm: /missing: No such file\n")
        }
    }

    @Test("seq prints integer, negative, decimal and padded sequences")
    func seq() throws {
        try withGuest { guest in
            expect(guest, "seq 3", "1\n2\n3\n")
            expect(guest, "seq 2 4", "2\n3\n4\n")
            expect(guest, "seq 1 2 7", "1\n3\n5\n7\n")
            expect(guest, "seq 1 2 8", "1\n3\n5\n7\n")
            expect(guest, "seq 5 -2 1", "5\n3\n1\n")
            expect(guest, "seq 3 -1 1", "3\n2\n1\n")
            expect(guest, "seq -1 1", "-1\n0\n1\n")
            expect(guest, "seq -3 -1", "-3\n-2\n-1\n")
            expect(guest, "seq 0", "")
            expect(guest, "seq 3 1", "")
            expect(guest, "seq 1 1", "1\n")
            expect(guest, "seq -s , 1 5", "1,2,3,4,5\n")
            expect(guest, "seq -s, 5", "1,2,3,4,5\n")
            expect(guest, "seq --separator=' ' 3", "1 2 3\n")
            expect(guest, "seq -s '' 3", "123\n")
            expect(guest, "seq -w 8 11", "08\n09\n10\n11\n")
            expect(guest, "seq --equal-width 1 3", "1\n2\n3\n")
            expect(guest, "seq -w 98 102", "098\n099\n100\n101\n102\n")
            expect(guest, "seq -w -s ' ' 9998 10001", "09998 09999 10000 10001\n")
            expect(guest, "seq -w -1 1", "-1\n00\n01\n")
            expect(guest, "seq 9998 10002", "9998\n9999\n10000\n10001\n10002\n")
            expect(guest, "seq 0 0.5 2", "0.0\n0.5\n1.0\n1.5\n2.0\n")
            expect(guest, "seq 1.5 3", "1.5\n2.5\n")
            expect(guest, "seq 1 0.25 1.5", "1.00\n1.25\n1.50\n")
            expect(guest, "seq 10 -2.5 0", "10.0\n7.5\n5.0\n2.5\n0.0\n")
            expect(guest, "seq 100000 | wc -l", "100000\n")
            expect(guest, "seq 100000 | tail -n 1", "100000\n")
            expect(guest, "seq 12345 | head -n 10001 | tail -n 2", "10000\n10001\n")
            expect(guest, "seq 1 0 2", "", status: 1,
                errors: "seq: invalid Zero increment value: '0'\n")
            expect(guest, "seq x", "", status: 1,
                errors: "seq: invalid floating point argument: 'x'\n")
            expect(
                guest, "seq", "", status: 1,
                errors: "seq: missing operand\nTry 'seq --help' for more information.\n")
            expect(
                guest, "seq 1 2 3 4",
                "", status: 1,
                errors: "seq: extra operand '4'\nTry 'seq --help' for more information.\n")
        }
    }

    @Test("uniq filters, counts and compares partial lines")
    func uniq() throws {
        try withGuest(files: Fixture.files.merging( ["/fields": "x 1\ny 1\nz 2\n"]) { $1 }) { guest in
            expect(guest, "uniq /repeats", "a\nb\nB\nc\n")
            expect(
                guest, "uniq -c /repeats",
                "      1 a\n      2 b\n      1 B\n      3 c\n")
            expect(guest, "uniq --count /left", "      1 a\n      1 c\n      1 e\n")
            expect(guest, "uniq -d /repeats", "b\nc\n")
            expect(guest, "uniq -u /repeats", "a\nB\n")
            expect(guest, "uniq -du /repeats", "")
            expect(guest, "uniq -cd /repeats", "      2 b\n      3 c\n")
            expect(guest, "uniq -i /repeats", "a\nb\nc\n")
            expect(guest, "uniq -ic /repeats", "      1 a\n      3 b\n      3 c\n")
            expect(guest, "uniq -f 1 /fields", "x 1\nz 2\n")
            expect(guest, "uniq -f1 -c /fields", "      2 x 1\n      1 z 2\n")
            expect(guest, "uniq -s 1 /fields", "x 1\nz 2\n")
            expect(guest, "uniq --skip-chars=2 /fields", "x 1\nz 2\n")
            expect(guest, "uniq -w 1 /repeats", "a\nb\nB\nc\n")
            expect(guest, "uniq -w 0 /repeats", "a\n")
            expect(guest, "uniq -i -w 1 -c /repeats", "      1 a\n      3 b\n      3 c\n")
            expect(guest, "cat /repeats | uniq -c", "      1 a\n      2 b\n      1 B\n      3 c\n")
            expect(guest, "uniq - < /repeats", "a\nb\nB\nc\n")
            expect(guest, "uniq /repeats /unique-output", "")
            expectFile(guest, "/unique-output", "a\nb\nB\nc\n")
            expect(guest, "uniq -c /repeats /unique-output", "")
            expectFile(guest, "/unique-output", "      1 a\n      2 b\n      1 B\n      3 c\n")
            expect(guest, "echo -n | uniq", "")
            expect(guest, "uniq /missing", "", status: 1, errors: "uniq: /missing: No such file\n")
            expect(guest, "uniq -f x /repeats", "", status: 1,
                errors: "uniq: x: invalid number of fields to skip\n")
            expect(
                guest, "uniq /a /b /c",
                "", status: 1,
                errors: "uniq: extra operand '/c'\nTry 'uniq --help' for more information.\n")
        }
    }

    @Test("tr translates, deletes, squeezes and complements sets")
    func tr() throws {
        try withGuest(files: Fixture.files) { guest in
            expect(guest, "echo banana | tr a-z A-Z", "BANANA\n")
            expect(guest, "echo banana | tr an xy", "bxyxyx\n")
            expect(guest, "echo abab | tr ab ba", "baba\n")
            expect(guest, "echo abcd | tr a-d b-da", "bcda\n")
            expect(guest, "echo hello | tr el x", "hxxxo\n")
            expect(guest, "echo hello | tr -t el x", "hxllo\n")
            expect(guest, "echo hello | tr --truncate-set1 abcdefgh xy", "hello\n")
            expect(guest, "echo abc | tr abc 'x[y*]'", "xyy\n")
            expect(guest, "echo abcd | tr abcd '[x*2]yz'", "xxyz\n")
            expect(guest, "echo abc | tr -d b", "ac\n")
            expect(guest, "echo abc | tr --delete a-b", "c\n")
            expect(guest, "echo aaabbbccc | tr -s ab", "abccc\n")
            expect(guest, "echo aaabbbccc | tr --squeeze-repeats a-c", "abc\n")
            expect(guest, "echo 'a  b   c' | tr -s ' ' '\\n'", "a\nb\nc\n")
            expect(guest, "echo aabbcc | tr -s ab xy", "xycc\n")
            expect(guest, "echo aabbcc | tr -ds a b", "bcc\n")
            expect(guest, "echo hello | tr -c 'a-z\\n' X", "hello\n")
            expect(guest, "echo 'a1 b2' | tr -c 'a-z\\n' _", "a__b_\n")
            expect(guest, "echo 'a1 b2' | tr -C 'a-z\\n' _", "a__b_\n")
            expect(guest, "echo a1b2 | tr -cd '[:digit:]\\n'", "12\n")
            expect(guest, "echo 'ab  12' | tr -cs '[:alnum:]\\n' -", "ab-12\n")
            expect(guest, "echo 'Hello World' | tr '[:upper:]' '[:lower:]'", "hello world\n")
            expect(guest, "echo 'Hello World' | tr '[:lower:]' '[:upper:]'", "HELLO WORLD\n")
            expect(guest, "echo 'a1 B2.' | tr -d '[:digit:]'", "a B.\n")
            expect(guest, "echo 'a1 B2.' | tr -d '[:alpha:]'", "1 2.\n")
            expect(guest, "echo 'a1 B2.' | tr -d '[:alnum:]'", " .\n")
            expect(guest, "echo 'a1 B2.' | tr -d '[:punct:]'", "a1 B2\n")
            expect(guest, "echo 'a1 B2.' | tr -d '[:space:]'", "a1B2.")
            expect(guest, "echo 'a1 B2.' | tr -d '[:blank:]'", "a1B2.\n")
            expect(guest, "echo 'fz09' | tr -d '[:xdigit:]'", "z\n")
            expect(guest, "echo 'a b' | tr -d '[:graph:]'", " \n")
            expect(guest, "echo 'a b' | tr -d '[:print:]'", "\n")
            expect(guest, "echo -e 'a\\tb' | tr -d '[:cntrl:]'", "ab")
            expect(guest, "echo -e 'a\\tb' | tr '\\t' ' '", "a b\n")
            expect(guest, "echo abc | tr a-c '\\101-\\103'", "ABC\n")
            expect(guest, "echo 'a\\b' | tr '\\\\' /", "a/b\n")
            expect(guest, "echo a-b | tr '\\-' _", "a_b\n")
            expect(guest, "echo a-b | tr -- -a _x", "x_b\n")
            expect(guest, "echo héllo | tr é e", "hello\n")
            expect(guest, "echo hello | tr e é", "héllo\n")
            expect(guest, "echo héllo | tr -d é", "hllo\n")
            expect(guest, "echo héllo | tr -cd 'a-z\\n'", "hllo\n")
            expect(guest, "tr a-c A-C < /left", "A\nC\ne\n")
            expect(
                guest, "tr", "", status: 1,
                errors: "tr: missing operand\nTry 'tr --help' for more information.\n")
            expect(
                guest, "echo x | tr a",
                "", status: 1,
                errors: "tr: missing operand after 'a'\nTwo strings must be given when translating.\n"
                    + "Try 'tr --help' for more information.\n")
            expect(guest, "echo x | tr '[:bogus:]' a", "", status: 1,
                errors: "tr: invalid character class 'bogus'\n")
            expect(
                guest, "echo x | tr z-a b",
                "", status: 1,
                errors: "tr: range-endpoints of 'z-a' are in reverse collating sequence order\n")
        }
    }
}
