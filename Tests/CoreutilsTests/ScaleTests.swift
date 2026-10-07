import Foundation
import Testing

/// Inputs of 100,000 lines: far more than one 64 KiB pipe buffer between
/// stages, and far more work than one event-loop drain. Every result is
/// checked exactly. Each case prints its wall time, because the guest VM is
/// slow in a debug build and these are the slowest tests in the package.
@Suite("scale")
struct ScaleTests {
    private static let count = 100_000

    private static func text(_ lines: [String]) -> String {
        lines.joined(separator: "\n") + "\n"
    }

    /// Runs `command`, reports how long it took, and checks its output.
    private static func timed(
        _ guest: Guest,
        _ command: String,
        _ output: String,
        sourceLocation: SourceLocation = #_sourceLocation
    ) {
        let start = Date()
        let result = guest.run(command)
        let seconds = String(format: "%.1f", Date().timeIntervalSince(start))
        print("scale: \(seconds)s  \(command)")
        let passed = result == CommandResult(output: output, errors: "", status: 0)
        guest.check(
            passed,
            "\(command)\n  status \(result.status), errors \(result.errors.debugDescription), "
                + "\(result.output.utf8.count) bytes of output, expected \(output.utf8.count)",
            sourceLocation: sourceLocation)
    }

    @Test("seq 1 100000 | sort -rn | head -3")
    func numericSortOfASequence() throws {
        try withGuest { guest in
            Self.timed(guest, "seq 1 100000 | sort -rn | head -3", "100000\n99999\n99998\n")
        }
    }

    @Test("sorts 100,000 lines of text in random order")
    func textSort() throws {
        var generator = Scramble(state: 0x7E57)
        let words = ["amber", "birch", "cedar", "delta", "ember", "fjord", "grove", "heron"]
        let lines = (0..<Self.count).map { index -> String in
            let word = words[generator.next() % words.count]
            return "\(word)-\(generator.next() % 50_000) \(index % 7)"
        }
        let ascending = lines.sorted { Array($0.utf8).lexicographicallyPrecedes(Array($1.utf8)) }
        let distinct = ascending.enumerated().filter { $0.offset == 0 || ascending[$0.offset - 1] != $0.element }
            .map(\.element)

        try withGuest(files: ["/words": Self.text(lines)]) { guest in
            Self.timed(guest, "sort /words", Self.text(ascending))
            Self.timed(guest, "cat /words | sort -r", Self.text(ascending.reversed()))
            Self.timed(guest, "sort -u /words", Self.text(distinct))
        }
    }

    @Test("passes 100,000 lines through cat, wc -l, tac, uniq and nl in pipelines")
    func hundredThousandLinePipelines() throws {
        // Every line appears twice in a row, so uniq has work to do.
        let lines = (0..<Self.count).map { "record \($0 / 2)" }
        let input = Self.text(lines)
        let distinct = (0..<Self.count / 2).map { "record \($0)" }
        let numbered = Self.text(lines.enumerated().map { index, line in
            let digits = String(index + 1)
            return String(repeating: " ", count: max(0, 6 - digits.count)) + digits + "\t" + line
        })

        try withGuest(files: ["/records": input]) { guest in
            Self.timed(guest, "cat /records | wc -l", "100000\n")
            Self.timed(guest, "cat /records | tac | tac", input)
            Self.timed(guest, "cat /records | uniq", Self.text(distinct))
            Self.timed(guest, "cat /records | nl", numbered)
            Self.timed(guest, "cat /records | tac | uniq | nl | tail -n 1", " 50000\trecord 0\n")
        }
    }
}
