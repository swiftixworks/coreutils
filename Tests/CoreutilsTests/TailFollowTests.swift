import Testing

/// `tail -f`: a named file is polled every sleep interval of logical time and
/// what was appended is printed; standard input is simply read to its end.
@Suite("tail -f")
struct TailFollowTests {
    @Test("prints what is appended to a growing file until interrupted")
    func growingFile() throws {
        try withGuest(files: ["/log": "one\ntwo\n"]) { guest in
            guest.type("tail -f -n 1 /log > /seen 2> /errors\n")
            expectFile(guest, "/seen", "two\n")

            // Nothing happens until the one-second interval has passed.
            guest.append("/log", "three\n")
            expectFile(guest, "/seen", "two\n")
            guest.advance(by: 0.5)
            expectFile(guest, "/seen", "two\n")
            guest.advance(by: 0.5)
            expectFile(guest, "/seen", "two\nthree\n")

            // Several appends between two looks arrive together, in order,
            // including a line that is still being written.
            guest.append("/log", "four\n")
            guest.append("/log", "five\nsi")
            guest.advance(by: 1)
            expectFile(guest, "/seen", "two\nthree\nfour\nfive\nsi")
            guest.append("/log", "x\n")
            guest.advance(by: 1)
            expectFile(guest, "/seen", "two\nthree\nfour\nfive\nsix\n")

            // An idle file costs nothing and prints nothing.
            guest.advance(by: 30)
            expectFile(guest, "/seen", "two\nthree\nfour\nfive\nsix\n")
            expectFile(guest, "/errors", "")

            // Ctrl-C ends the follower and the shell takes the next command.
            guest.type("\u{03}")
            guest.append("/log", "seven\n")
            guest.advance(by: 5)
            expectFile(guest, "/seen", "two\nthree\nfour\nfive\nsix\n")
            expect(guest, "echo back", "back\n")
        }
    }

    @Test("starts again from the top of a file that was truncated")
    func truncatedFile() throws {
        try withGuest(files: ["/log": "first run\nsecond line\n"]) { guest in
            guest.type("tail -f /log > /seen 2> /errors\n")
            expectFile(guest, "/seen", "first run\nsecond line\n")
            guest.write("/log", "rotated\n")
            guest.advance(by: 1)
            expectFile(guest, "/seen", "first run\nsecond line\nrotated\n")
            expectFile(guest, "/errors", "tail: /log: file truncated\n")
            guest.append("/log", "more\n")
            guest.advance(by: 1)
            expectFile(guest, "/seen", "first run\nsecond line\nrotated\nmore\n")
            guest.type("\u{03}")
            expect(guest, "echo back", "back\n")
        }
    }

    @Test("honours -s and --sleep-interval, whole and fractional")
    func sleepInterval() throws {
        try withGuest(files: ["/log": "start\n"]) { guest in
            guest.type("tail -f -s 3 -n 0 /log > /seen\n")
            guest.append("/log", "late\n")
            guest.advance(by: 2.5)
            expectFile(guest, "/seen", "")
            guest.advance(by: 0.5)
            expectFile(guest, "/seen", "late\n")
            guest.type("\u{03}")

            guest.type("tail --follow --sleep-interval=0.25 -n 0 /log > /seen\n")
            guest.append("/log", "soon\n")
            guest.advance(by: 0.2)
            expectFile(guest, "/seen", "")
            guest.advance(by: 0.05)
            expectFile(guest, "/seen", "soon\n")
            guest.type("\u{03}")
            expect(guest, "echo back", "back\n")
        }
    }

    @Test("follows several files and says which one each piece came from")
    func severalFiles() throws {
        try withGuest(files: ["/a": "a1\n", "/b": "b1\n"]) { guest in
            guest.type("tail -f -n 1 /a /b > /seen\n")
            expectFile(guest, "/seen", "==> /a <==\na1\n\n==> /b <==\nb1\n")
            guest.append("/b", "b2\n")
            guest.advance(by: 1)
            guest.append("/a", "a2\n")
            guest.advance(by: 1)
            guest.append("/a", "a3\n")
            guest.advance(by: 1)
            expectFile(
                guest, "/seen",
                "==> /a <==\na1\n\n==> /b <==\nb1\nb2\n\n==> /a <==\na2\na3\n")
            guest.type("\u{03}")

            guest.type("tail -f -q -c 1 /a /b > /seen\n")
            guest.append("/b", "b3\n")
            guest.advance(by: 1)
            expectFile(guest, "/seen", "\n\nb3\n")
            guest.type("\u{03}")
            expect(guest, "echo back", "back\n")
        }
    }

    @Test("lets other commands run while it waits, and feeds a pipe")
    func concurrentWriter() throws {
        try withGuest(files: ["/log": ""]) { guest in
            // A background follower in a pipeline, written to by later
            // foreground commands of the same shell.
            guest.type("tail -f /log | cat > /seen &\n")
            guest.type("echo first >> /log\n")
            guest.advance(by: 1)
            guest.type("seq 3 >> /log\n")
            guest.advance(by: 1)
            expectFile(guest, "/log", "first\n1\n2\n3\n")
            expectFile(guest, "/seen", "first\n1\n2\n3\n")
        }
    }

    @Test("reads standard input to its end and then stops")
    func standardInput() throws {
        try withGuest(files: Fixture.files) { guest in
            // A pipe: -f adds nothing, and the command ends with its input.
            expect(guest, "cat /left | tail -f -n 2", "c\ne\n")
            expect(guest, "cat /left | tail --follow -n 1 -", "e\n")
            expect(guest, "seq 20000 | tail -f -c 6", "20000\n")
            expect(guest, "tail -f -n 1 < /left", "e\n")

            // A terminal: lines are awaited as they are typed, nothing is
            // printed before end of input, and Ctrl-D ends the command.
            guest.type("tail -f -n 2 > /typed\n")
            guest.type("one\n")
            guest.type("two\n")
            guest.advance(by: 10)
            expectFile(guest, "/typed", "")
            guest.type("three\n")
            guest.type("\u{04}")
            expectFile(guest, "/typed", "two\nthree\n")
            expect(guest, "echo back", "back\n")
        }
    }

    @Test("a file that cannot be opened is reported and not followed")
    func missingFile() throws {
        try withGuest(files: Fixture.files) { guest in
            // With nothing left to follow the command ends at once.
            expect(guest, "tail -f /missing", "", status: 1, errors: "tail: /missing: No such file\n")
        }
    }
}
