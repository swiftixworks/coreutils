import Testing

/// `sort` orderings, keys, and modes, executed from the packaged image.
@Suite("sort")
struct SortTests {
    private static let files: [String: String] = Fixture.files.merging([
        "/sizes": "2K\n1G\n512\n3M\n10K\n-1K\n",
        "/indented": "  b\n a\nc\n",
        "/mixed": "b\nA\na\nB\n",
    ]) { $1 }

    @Test("orders whole lines by byte value, reversed, unique and folded")
    func wholeLines() throws {
        try withGuest(files: Self.files) { guest in
            expect(guest, "sort /fruits", "apple\napple\nbanana\ncherry\n")
            expect(guest, "sort -r /fruits", "cherry\nbanana\napple\napple\n")
            expect(guest, "sort --reverse /fruits", "cherry\nbanana\napple\napple\n")
            expect(guest, "sort -u /fruits", "apple\nbanana\ncherry\n")
            expect(guest, "sort --unique /fruits", "apple\nbanana\ncherry\n")
            expect(guest, "sort -ru /fruits", "cherry\nbanana\napple\n")
            expect(guest, "sort -u -r /fruits", "cherry\nbanana\napple\n")
            expect(guest, "sort /mixed", "A\nB\na\nb\n")
            expect(guest, "sort -f /mixed", "A\na\nB\nb\n")
            expect(guest, "sort --ignore-case /mixed", "A\na\nB\nb\n")
            expect(guest, "sort -fu /mixed", "A\nb\n")
            expect(guest, "sort -fs /mixed", "A\na\nb\nB\n")
            expect(guest, "sort -fr /mixed", "b\nB\na\nA\n")
            expect(guest, "sort /indented", "  b\n a\nc\n")
            expect(guest, "sort -b /indented", " a\n  b\nc\n")
            expect(guest, "sort --ignore-leading-blanks /indented", " a\n  b\nc\n")
            expect(guest, "sort /blanks", "\n\n\n\none\nthree\ntwo\tx\n")
            expect(guest, "sort /left /right", "a\nb\nc\nc\nd\ne\n")
            expect(guest, "sort -m /left /right", "a\nb\nc\nc\nd\ne\n")
            expect(guest, "sort -u /left /right", "a\nb\nc\nd\ne\n")
            expect(guest, "cat /fruits | sort", "apple\napple\nbanana\ncherry\n")
            expect(guest, "cat /fruits | sort -r | head -n 1", "cherry\n")
            expect(guest, "sort - /left < /right", "a\nb\nc\nc\nd\ne\n")
            expect(guest, "echo -n | sort", "")
            expect(guest, "echo -n b | sort", "b\n")
        }
    }

    @Test("orders numbers with -n, -rn and -h")
    func numeric() throws {
        try withGuest(files: Self.files) { guest in
            expect(guest, "sort -n /numbers", "-10\n-3\n0\nabc\n2.25\n2.5\n9\n10\n")
            expect(guest, "sort --numeric-sort /numbers", "-10\n-3\n0\nabc\n2.25\n2.5\n9\n10\n")
            expect(guest, "sort -rn /numbers", "10\n9\n2.5\n2.25\nabc\n0\n-3\n-10\n")
            expect(guest, "sort -nr /numbers", "10\n9\n2.5\n2.25\nabc\n0\n-3\n-10\n")
            expect(guest, "sort -n -r /numbers", "10\n9\n2.5\n2.25\nabc\n0\n-3\n-10\n")
            expect(guest, "sort -nu /numbers", "-10\n-3\nabc\n2.25\n2.5\n9\n10\n")
            expect(guest, "sort /numbers", "-10\n-3\n0\n10\n2.25\n2.5\n9\nabc\n")
            expect(guest, "seq 12 | sort", "1\n10\n11\n12\n2\n3\n4\n5\n6\n7\n8\n9\n")
            expect(guest, "seq 12 | sort -n | tail -n 3", "10\n11\n12\n")
            expect(guest, "seq 12 | sort -rn | head -n 3", "12\n11\n10\n")
            expect(guest, "seq 8 12 | sort -r", "9\n8\n12\n11\n10\n")
            expect(guest, "echo -e '007\\n7.0\\n 7\\n8\\n6.99' | sort -n", "6.99\n 7\n007\n7.0\n8\n")
            expect(guest, "echo -e '-0.5\\n-0.25\\n-1\\n0.5' | sort -n", "-1\n-0.5\n-0.25\n0.5\n")
            expect(guest, "sort -h /sizes", "-1K\n512\n2K\n10K\n3M\n1G\n")
            expect(guest, "sort --human-numeric-sort -r /sizes", "1G\n3M\n10K\n2K\n512\n-1K\n")
        }
    }

    @Test("sorts by field keys with -k and -t")
    func keys() throws {
        try withGuest(files: Self.files) { guest in
            expect(guest, "sort -k2 /pairs", "apple 10\nCherry 2\napple 2\nbanana 3\n")
            expect(guest, "sort -k 2 /pairs", "apple 10\nCherry 2\napple 2\nbanana 3\n")
            expect(guest, "sort --key=2 /pairs", "apple 10\nCherry 2\napple 2\nbanana 3\n")
            expect(guest, "sort -k2n /pairs", "Cherry 2\napple 2\nbanana 3\napple 10\n")
            expect(guest, "sort -k2 -n /pairs", "Cherry 2\napple 2\nbanana 3\napple 10\n")
            expect(guest, "sort -k2,2n /pairs", "Cherry 2\napple 2\nbanana 3\napple 10\n")
            expect(guest, "sort -k2,2nr /pairs", "apple 10\nbanana 3\napple 2\nCherry 2\n")
            expect(guest, "sort -rn -k2 /pairs", "apple 10\nbanana 3\napple 2\nCherry 2\n")
            expect(guest, "sort -k1,1 /pairs", "Cherry 2\napple 10\napple 2\nbanana 3\n")
            expect(guest, "sort -k1,1 -s /pairs", "Cherry 2\napple 10\napple 2\nbanana 3\n")
            expect(guest, "sort -k1,1f /pairs", "apple 10\napple 2\nbanana 3\nCherry 2\n")
            expect(guest, "sort -f -k1,1 -k2n /pairs", "apple 2\napple 10\nbanana 3\nCherry 2\n")
            expect(guest, "sort -k1,1 -k2n /pairs", "Cherry 2\napple 2\napple 10\nbanana 3\n")
            expect(guest, "sort -k1.2,1.3 /pairs", "banana 3\nCherry 2\napple 10\napple 2\n")
            expect(guest, "sort -t: -k2 /table", "d:1:x\nb:1:y\nc:2:x\na:3:x\n")
            expect(guest, "sort -t : -k 2,2 /table", "b:1:y\nd:1:x\nc:2:x\na:3:x\n")
            expect(guest, "sort --field-separator=: -k2,2 -s /table", "b:1:y\nd:1:x\nc:2:x\na:3:x\n")
            expect(guest, "sort -t: -k2,2 -rs /table", "a:3:x\nc:2:x\nb:1:y\nd:1:x\n")
            expect(guest, "sort -t: -k2,2 -r /table", "a:3:x\nc:2:x\nd:1:x\nb:1:y\n")
            expect(guest, "sort -t: -k2,2 -u /table", "b:1:y\nc:2:x\na:3:x\n")
            expect(guest, "sort -t: -k3,3 -k2,2n /table", "d:1:x\nc:2:x\na:3:x\nb:1:y\n")
            expect(guest, "sort -t: -k3r -k1,1r /table", "b:1:y\nd:1:x\nc:2:x\na:3:x\n")
            expect(guest, "sort -b -k2 /pairs", "apple 10\nCherry 2\napple 2\nbanana 3\n")
            expect(guest, "sort -k2b,2 /pairs", "apple 10\nCherry 2\napple 2\nbanana 3\n")
            expect(guest, "sort -t: -k9 /table", "a:3:x\nb:1:y\nc:2:x\nd:1:x\n")
        }
    }

    @Test("writes to -o, checks with -c, and reports misuse with status 2")
    func modes() throws {
        try withGuest(files: Self.files) { guest in
            expect(guest, "sort -o /sorted /fruits", "")
            expectFile(guest, "/sorted", "apple\napple\nbanana\ncherry\n")
            expect(guest, "sort --output=/sorted -r /fruits", "")
            expectFile(guest, "/sorted", "cherry\nbanana\napple\napple\n")
            expect(guest, "sort -u -o /sorted /sorted", "")
            expectFile(guest, "/sorted", "apple\nbanana\ncherry\n")
            expect(guest, "sort -c /sorted", "")
            expect(guest, "sort -c /fruits", "", status: 1,
                errors: "sort: /fruits:2: disorder: apple\n")
            expect(guest, "sort --check /fruits", "", status: 1,
                errors: "sort: /fruits:2: disorder: apple\n")
            expect(guest, "sort -C /fruits", "", status: 1)
            expect(guest, "sort -cr /sorted", "", status: 1,
                errors: "sort: /sorted:2: disorder: banana\n")
            expect(guest, "sort /fruits | sort -c", "")
            expect(guest, "sort /fruits | sort -cu", "", status: 1,
                errors: "sort: -:2: disorder: apple\n")
            expect(guest, "seq 12 | sort -c", "", status: 1,
                errors: "sort: -:10: disorder: 10\n")
            expect(guest, "seq 12 | sort -cn", "")
            expect(
                guest, "sort -x /fruits",
                "", status: 2,
                errors: "sort: invalid option -- 'x'\nTry 'sort --help' for more information.\n")
            expect(
                guest, "sort -k",
                "", status: 2,
                errors: "sort: option requires an argument -- 'k'\nTry 'sort --help' for more information.\n")
            expect(guest, "sort -t ab /fruits", "", status: 2,
                errors: "sort: multi-character tab 'ab'\n")
            expect(
                guest, "sort /fruits /missing", "", status: 2,
                errors: "sort: /missing: No such file\n")
            expect(
                guest, "sort -k1,1 -k2r /pairs",
                "", status: 2,
                errors: "sort: keys that mix ascending and descending order are not supported\n")
        }
    }

    @Test("is stable and correct on a few hundred lines in random order")
    func manyLines() throws {
        var generator = Scramble(state: 0x5EED)
        let lines = (0..<240).map { "key\(generator.next() % 60) row\($0)" }
        let byBytes = lines.sorted { Array($0.utf8).lexicographicallyPrecedes(Array($1.utf8)) }
        // -s keeps input order within a key; the row numbers prove it.
        func keyOf(_ line: String) -> [UInt8] { Array(line.split(separator: " ")[0].utf8) }
        let stable = lines.enumerated().sorted {
            let left = keyOf($0.element)
            let right = keyOf($1.element)
            if left != right { return left.lexicographicallyPrecedes(right) }
            return $0.offset < $1.offset
        }.map(\.element)

        // -k1.4n reads the number after "key"; equal numbers fall back to bytes.
        func numberOf(_ line: String) -> Int { Int(line.split(separator: " ")[0].dropFirst(3))! }
        let numeric = lines.sorted {
            let left = numberOf($0)
            let right = numberOf($1)
            if left != right { return left < right }
            return Array($0.utf8).lexicographicallyPrecedes(Array($1.utf8))
        }

        try withGuest(files: ["/rows": Self.text(lines)]) { guest in
            expect(guest, "sort /rows", Self.text(byBytes))
            expect(guest, "sort -r /rows", Self.text(byBytes.reversed()))
            expect(guest, "sort -s -k1,1 /rows", Self.text(stable))
            expect(guest, "sort -k1.4n /rows", Self.text(numeric))
            expect(guest, "sort -k1.4nr /rows", Self.text(numeric.reversed()))
            expect(guest, "cat /rows | sort | sort -c", "")
        }
    }

    @Test("orders a few hundred shuffled numbers with -n and -rn")
    func manyNumbers() throws {
        var generator = Scramble(state: 0xC0FFEE)
        var values = Array(1...300).map { $0 * 7 - 1_000 }
        for index in stride(from: values.count - 1, to: 0, by: -1) {
            values.swapAt(index, generator.next() % (index + 1))
        }
        let input = values.map { "\($0)\n" }.joined()
        let ascending = values.sorted().map { "\($0)\n" }.joined()
        let descending = values.sorted(by: >).map { "\($0)\n" }.joined()

        try withGuest(files: ["/values": input]) { guest in
            expect(guest, "sort -n /values", ascending)
            expect(guest, "sort -rn /values", descending)
            expect(guest, "cat /values | sort -n | sort -cn", "")
            expect(guest, "sort -n /values | head -n 2", "-993\n-986\n")
            expect(guest, "sort -rn /values | head -n 2", "1100\n1093\n")
        }
    }

    @Test("sorts 5,000 lines that arrive nearly ordered")
    func fiveThousandOrderedLines() throws {
        let lines = (0..<5_000).map { String(format: "entry %05d", $0) }
        // Two displaced lines keep this a real sort, not a no-op.
        var displaced = lines
        displaced.swapAt(1_200, 3_700)
        let input = Self.text(displaced)

        try withGuest(files: ["/entries": input]) { guest in
            expect(guest, "sort -o /ordered /entries", "")
            expectFile(guest, "/ordered", Self.text(lines))
            expect(guest, "sort -c /ordered", "")
            expect(guest, "sort -u /ordered | wc -l", "5000\n")
        }
    }

    private static func text(_ lines: [String]) -> String {
        lines.map { $0 + "\n" }.joined()
    }
}

/// A tiny deterministic generator so fixtures are identical on every run.
struct Scramble {
    var state: UInt64

    mutating func next() -> Int {
        state = state &* 6_364_136_223_846_793_005 &+ 1_442_695_040_888_963_407
        return Int(state >> 33)
    }
}
