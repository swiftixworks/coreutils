import Testing

/// Behaviour that spans commands: standard input from pipes, redirections and
/// the terminal, large inputs, and the shared option parser.
@Suite("pipelines and shared behaviour")
struct PipelineTests {
    @Test("every option-parsing command prints usage for --help and exits 0")
    func help() throws {
        try withGuest(files: Fixture.files) { guest in
            for name in CoreutilsArchive.commandNames where !["echo", "true", "false"].contains(name) {
                let result = guest.run("\(name) --help")
                guest.check(
                    result.status == 0 && result.errors == ""
                        && result.output.hasPrefix("Usage: \(name) "),
                    "\(name) --help: \(result)")
            }
            expect(guest, "echo --help", "--help\n")
        }
    }

    @Test("an unknown option is reported on standard error and fails without reading input")
    func unknownOptions() throws {
        try withGuest(files: Fixture.files) { guest in
            for name in ["cat", "comm", "fold", "head", "nl", "paste", "rev", "seq", "tac", "tail", "tr", "uniq", "wc"] {
                expect(
                    guest, "\(name) -Z",
                    "", status: 1,
                    errors: "\(name): invalid option -- 'Z'\nTry '\(name) --help' for more information.\n")
            }
            expect(
                guest, "sort -Z",
                "", status: 2,
                errors: "sort: invalid option -- 'Z'\nTry 'sort --help' for more information.\n")
            expect(
                guest, "wc --lines=3 /left",
                "", status: 1,
                errors: "wc: option '--lines' doesn't allow an argument\nTry 'wc --help' for more information.\n")
            expect(
                guest, "head --lines",
                "", status: 1,
                errors: "head: option '--lines' requires an argument\nTry 'head --help' for more information.\n")
        }
    }

    @Test("-- ends options and - names standard input")
    func terminatorAndDash() throws {
        try withGuest(files: Fixture.files.merging(["/-n": "dash file\n"]) { $1 }) { guest in
            expect(guest, "cat -- /-n", "dash file\n")
            expect(guest, "wc -l -- /-n", "1 /-n\n")
            expect(guest, "sort -- /left", "a\nc\ne\n")
            expect(guest, "cat /left | sort -r -", "e\nc\na\n")
            expect(guest, "cat /left | head -n 1 -", "a\n")
            expect(guest, "cat /left | uniq -c -", "      1 a\n      1 c\n      1 e\n")
            expect(guest, "cat /left | nl -", "     1\ta\n     2\tc\n     3\te\n")
            expect(guest, "cat /left | fold -w 1 -", "a\nc\ne\n")
        }
    }

    @Test("classic multi-stage pipelines produce the conventional result")
    func classicPipelines() throws {
        let text = "the quick brown fox\njumps over the lazy dog\nThe dog barks; the fox runs.\n"
        try withGuest(files: Fixture.files.merging(["/text": text]) { $1 }) { guest in
            expect(
                guest,
                "cat /text | tr -cs 'A-Za-z' '\\n' | tr A-Z a-z | sort | uniq -c | sort -rn | head -n 3",
                "      4 the\n      2 fox\n      2 dog\n")
            expect(guest, "sort /fruits | uniq -c | sort -rn | head -n 1", "      2 apple\n")
            expect(guest, "sort /fruits | uniq | wc -l", "3\n")
            expect(guest, "seq 10 | tac | head -n 3 | paste -s -d ,", "10,9,8\n")
            expect(guest, "seq 20 | tail -n 5 | head -n 2", "16\n17\n")
            expect(guest, "cat /left /right | sort | uniq -d", "c\n")
            expect(guest, "seq 3 | rev | cat -n | tac", "     3\t3\n     2\t2\n     1\t1\n")
            expect(guest, "echo 'b a c' | tr ' ' '\\n' | sort | paste -s -d ' '", "a b c\n")
            expect(guest, "seq 1 2 9 | nl -w 1 -s = | fold -w 3 | wc -l", "5\n")
            expect(guest, "comm -12 /left /right | wc -c", "2\n")
        }
    }

    @Test("reads typed terminal input until end of file")
    func terminalInput() throws {
        try withGuest { guest in
            guest.type("cat > /typed\n")
            guest.type("first line\n")
            guest.type("second line\n")
            guest.type("\u{04}")
            expectFile(guest, "/typed", "first line\nsecond line\n")

            guest.type("sort -r > /typed\n")
            guest.type("a\nc\nb\n")
            guest.type("\u{04}")
            expectFile(guest, "/typed", "c\nb\na\n")

            guest.type("wc -l > /typed\n")
            guest.type("one\ntwo\n\u{04}")
            expectFile(guest, "/typed", "2\n")
        }
    }

    @Test("counts 5,000 lines and their words, characters and width")
    func fiveThousandLinesThroughWc() throws {
        let lines = (1...5_000).map { "row \($0) of the fixture\n" }
        let text = lines.joined()
        try withGuest(files: ["/rows": text]) { guest in
            let bytes = text.utf8.count
            expect(guest, "wc /rows", "  5000  25000 \(bytes) /rows\n")
            expect(guest, "wc -l /rows", "5000 /rows\n")
            expect(guest, "wc -w /rows", "25000 /rows\n")
            expect(guest, "wc -m /rows", "\(bytes) /rows\n")
            expect(guest, "wc -L /rows", "23 /rows\n")
            expect(guest, "cat /rows | wc -l", "5000\n")
            expect(guest, "head -n 4999 /rows | tail -n 1", "row 4999 of the fixture\n")
            expect(guest, "tail -n 2500 /rows | head -n 1", "row 2501 of the fixture\n")
            expect(guest, "tail -n +4000 /rows | wc -l", "1001\n")
            expect(guest, "tr a-z A-Z < /rows | tail -n 1", "ROW 5000 OF THE FIXTURE\n")
            expect(guest, "tr -d 'a-z ' < /rows | tail -n 1", "5000\n")
            expect(guest, "uniq /rows | wc -l", "5000\n")
            expect(guest, "tac /rows | head -n 1", "row 5000 of the fixture\n")
            expect(guest, "fold -w 80 /rows | wc -l", "5000\n")
        }
    }

    @Test("carries more than a pipe's capacity between two packaged commands")
    func largePipeTransfer() throws {
        // 200 KiB is three times the 64 KiB pipe: each writer has to wait for
        // its reader several times, and nothing may be lost on the way.
        let line = String(repeating: "x", count: 99) + "\n"
        let text = String(repeating: line, count: 2_048)
        try withGuest(files: ["/large": text]) { guest in
            expect(guest, "cat /large | wc -c", "204800\n")
            expect(guest, "cat /large | cat | wc -l", "2048\n")
            expect(guest, "cat < /large | wc -c", "204800\n")
            expect(guest, "tr x y < /large | tr y z | tail -c 5", "zzzz\n")
            expect(guest, "cat /large | head -c 150000 | wc -c", "150000\n")
            expect(guest, "cat /large | tail -n 1 | wc -c", "100\n")
        }
    }
}
