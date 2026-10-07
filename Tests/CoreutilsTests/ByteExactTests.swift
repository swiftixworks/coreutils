import Testing

/// `head -c`, `tail -c` and plain `cat` handle bytes, not characters: a count
/// may cut a UTF-8 sequence, and input that is not text passes through intact.
@Suite("byte-exact commands")
struct ByteExactTests {
    /// "héllo wörld\n": é is bytes 1...2 and ö bytes 8...9.
    private static let text = Array("héllo wörld\n".utf8)

    /// Every byte value twice over, then sequences that are invalid UTF-8.
    private static let binary: [UInt8] =
        (0...255).map { UInt8($0) } + (0...255).reversed().map { UInt8($0) }
        + [0xC3, 0x28, 0xE2, 0x82, 0xFF, 0x00, 0xFE, 0x0A, 0xF0, 0x9F]

    @Test("head -c and tail -c cut inside a multi-byte character")
    func multiByteText() throws {
        let text = Self.text
        try withGuest { guest in
            guest.write("/text", text)
            for count in 0...text.count + 1 {
                let head = Array(text.prefix(count))
                let tail = Array(text.suffix(count))
                expectBytes(guest, "head -c \(count) /text", head)
                expectBytes(guest, "tail -c \(count) /text", tail)
            }
            expectBytes(guest, "head -c 2 /text", [0x68, 0xC3])
            expectBytes(guest, "tail -c 5 /text", [0xB6, 0x72, 0x6C, 0x64, 0x0A])
            expectBytes(guest, "head -c -12 /text", [0x68, 0xC3])
            expectBytes(guest, "tail -c +3 /text", Array(text.dropFirst(2)))
            expectBytes(guest, "head --bytes=2 - < /text", [0x68, 0xC3])
            expectBytes(guest, "cat /text | head -c 2", [0x68, 0xC3])
            expectBytes(guest, "cat /text | tail -c 5", Array(text.suffix(5)))
            expectBytes(guest, "cat /text | head -c -12", [0x68, 0xC3])
            expectBytes(guest, "cat /text | tail -c +10", Array(text.dropFirst(9)))
            // Both halves of a split character still make the whole file.
            expectBytes(guest, "head -c 2 /text > /a; tail -c +3 /text > /b; cat /a /b", text)
        }
    }

    @Test("binary input passes through head -c, tail -c and cat unchanged")
    func binaryInput() throws {
        let binary = Self.binary
        try withGuest { guest in
            guest.write("/binary", binary)
            expectBytes(guest, "cat /binary", binary)
            expectBytes(guest, "cat < /binary", binary)
            expectBytes(guest, "cat /binary | cat", binary)
            expectBytes(guest, "cat /binary - /binary < /binary", binary + binary + binary)
            expectBytes(guest, "head -c 200 /binary", Array(binary.prefix(200)))
            expectBytes(guest, "head -c 1K /binary", binary)
            expectBytes(guest, "head -c -500 /binary", Array(binary.dropLast(500)))
            expectBytes(guest, "tail -c 300 /binary", Array(binary.suffix(300)))
            expectBytes(guest, "tail -c 10 /binary", Array(binary.suffix(10)))
            expectBytes(guest, "tail -c +257 /binary", Array(binary.dropFirst(256)))
            expectBytes(guest, "cat /binary | head -c 300 | tail -c 100", Array(binary[200..<300]))
            expectBytes(guest, "cat /binary | tail -c 266 | head -c 256", Array(binary[256..<512]))
            expectBytes(
                guest, "head -q -c 3 /binary /binary", Array(binary.prefix(3)) + Array(binary.prefix(3)))
            expectBytes(
                guest, "tail -c 2 /binary /binary",
                Array("==> /binary <==\n".utf8) + binary.suffix(2)
                    + Array("\n==> /binary <==\n".utf8) + binary.suffix(2))
            // The pieces of a split binary file reassemble exactly.
            expectBytes(guest, "head -c 259 /binary > /a; tail -c +260 /binary > /b; cat /a /b", binary)
        }
    }

    @Test("carries 300,000 arbitrary bytes through pipes without changing one")
    func largeBinaryThroughPipes() throws {
        var generator = Scramble(state: 0xB17E5)
        let blob = (0..<300_000).map { _ in UInt8(truncatingIfNeeded: generator.next()) }
        try withGuest { guest in
            guest.write("/blob", blob)
            expectBytes(guest, "cat /blob | cat", blob)
            expectBytes(guest, "cat /blob | head -c 170001", Array(blob.prefix(170_001)))
            expectBytes(guest, "cat /blob | tail -c 170001", Array(blob.suffix(170_001)))
            expectBytes(guest, "cat /blob | head -c -1 | tail -c +2", Array(blob.dropFirst().dropLast()))
            expectBytes(guest, "tail -c 100K /blob | head -c 5", Array(blob.suffix(102_400).prefix(5)))
        }
    }
}
