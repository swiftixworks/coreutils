import Testing

/// Standard error is separate from standard output: a failing command leaves
/// its results clean, and its exit status is the conventional one.
@Suite("diagnostics")
struct DiagnosticsTests {
    @Test("a usage error writes nothing to standard output")
    func usageErrors() throws {
        try withGuest(files: Fixture.files) { guest in
            expect(
                guest, "cat -x /left", "", status: 1,
                errors: "cat: invalid option -- 'x'\nTry 'cat --help' for more information.\n")
            expect(
                guest, "sort --bogus /left", "", status: 2,
                errors: "sort: unrecognized option '--bogus'\nTry 'sort --help' for more information.\n")
            expect(
                guest, "head -n", "", status: 1,
                errors: "head: option requires an argument -- 'n'\nTry 'head --help' for more information.\n")
            expect(
                guest, "seq", "", status: 1,
                errors: "seq: missing operand\nTry 'seq --help' for more information.\n")
            expect(guest, "tail -c x /left", "", status: 1, errors: "tail: invalid number of bytes: 'x'\n")
            expect(guest, "tail -f -s x /left", "", status: 1, errors: "tail: invalid number of seconds: 'x'\n")
        }
    }

    @Test("results already produced stay on standard output when a later step fails")
    func partialResults() throws {
        try withGuest(files: Fixture.files) { guest in
            expect(
                guest, "cat /left /missing /right", "a\nc\ne\nb\nc\nd\n", status: 1,
                errors: "cat: /missing: No such file\n")
            expect(
                guest, "head -c 2 /missing /left", "==> /left <==\na\n", status: 1,
                errors: "head: /missing: No such file\n")
            expect(
                guest, "tail -c 2 /left /missing", "==> /left <==\ne\n", status: 1,
                errors: "tail: /missing: No such file\n")
            expect(
                guest, "wc -l /missing /left", "3 /left\n3 total\n", status: 1,
                errors: "wc: /missing: No such file\n")
            expect(guest, "sort -c /fruits", "", status: 1, errors: "sort: /fruits:2: disorder: apple\n")
        }
    }

    @Test("a diagnostic does not enter the pipe to the next command")
    func pipelines() throws {
        try withGuest(files: Fixture.files) { guest in
            // The reader sees an empty stream, not the writer's message.
            expect(guest, "cat -x /left 2> /first | wc -c", "0\n")
            expectFile(
                guest, "/first", "cat: invalid option -- 'x'\nTry 'cat --help' for more information.\n")
            expect(guest, "seq x 2> /first | sort | wc -l", "0\n")
            expectFile(guest, "/first", "seq: invalid floating point argument: 'x'\n")
            expect(guest, "cat /left /missing 2> /first | tac", "e\nc\na\n")
            expectFile(guest, "/first", "cat: /missing: No such file\n")
            // The pipeline's status and captured errors are its last command's.
            expect(
                guest, "cat /left | sort -c -r", "", status: 1,
                errors: "sort: -:2: disorder: c\n")
            expect(
                guest, "cat /left | tr", "", status: 1,
                errors: "tr: missing operand\nTry 'tr --help' for more information.\n")
            // Left alone, a diagnostic reaches the terminal even when standard
            // output is a pipe or a file.
            guest.type("uniq -Z /left | wc -l > /count\n")
            expectFile(guest, "/count", "0\n")
            guest.check(
                guest.screen.contains("uniq: invalid option -- 'Z'\n"),
                "the terminal showed \(guest.screen.debugDescription)")
        }
    }

    @Test("standard error can be redirected on its own and merged on request")
    func redirection() throws {
        try withGuest(files: Fixture.files) { guest in
            guest.type("nl -b x /left > /out 2> /err\n")
            expectFile(guest, "/out", "")
            expectFile(guest, "/err", "nl: invalid body numbering style: 'x'\n")
            guest.type("cat /left /missing > /both 2>&1\n")
            expectFile(guest, "/both", "a\nc\ne\ncat: /missing: No such file\n")
            // --help is a result, not a diagnostic.
            guest.type("fold --help 2> /err > /out\n")
            expectFile(guest, "/err", "")
            guest.check(guest.read("/out").hasPrefix("Usage: fold "), "fold --help went to \(guest.read("/out"))")
        }
    }
}
