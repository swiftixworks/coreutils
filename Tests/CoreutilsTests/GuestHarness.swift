import Foundation
import Swiftix
import SwiftixGoRuntime
import SwiftixPackages
import Testing

/// What one shell command line produced.
struct CommandResult: Equatable {
    var output: String
    var errors: String
    var status: Int32
}

/// The checked-in package archive under test.
enum CoreutilsArchive {
    static let version = "1.1.0"

    static let commandNames = [
        "cat", "comm", "echo", "false", "fold", "head", "nl", "paste",
        "rev", "seq", "sort", "tac", "tail", "tr", "true", "uniq", "wc",
    ]

    static func load() throws -> PackageArchive {
        let root = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()  // CoreutilsTests
            .deletingLastPathComponent()  // Tests
            .deletingLastPathComponent()  // repository root
        let bytes = Array(try Data(contentsOf: root
            .appendingPathComponent("Artifacts/coreutils_\(version).pkg")))
        return try PackageArchive.decode(bytes)
    }
}

/// A fresh Swiftix kernel with the packaged commands installed in `/usr/bin`
/// and a shell attached to a pseudo-terminal.
///
/// Commands are resolved through `$PATH` exactly as on an installed system, so
/// every test runs the packaged Go images on the real Swiftix Go runtime, not
/// the shell's built-in fallbacks. Everything runs on the calling thread; the
/// event loop is driven to idle after each line, which leaves logical time
/// where it was, so a process that sleeps stays asleep until `advance(by:)`.
/// Create one with `withGuest`.
final class Guest {
    let loop = EventLoop()
    let kernel: Kernel
    let terminal = PseudoTerminal()
    private var screenBytes: [UInt8] = []
    fileprivate var failures: [GuestFailure] = []

    init(files: [String: String] = [:]) throws {
        let archive = try CoreutilsArchive.load()
        kernel = Kernel(loop: loop)
        kernel.spawn("install-coreutils") { context in
            _ = context.mkdir("/usr")
            _ = context.mkdir("/usr/bin")
            for entry in archive.files where entry.path.hasPrefix("/usr/bin/") {
                Guest.write(context, entry.path, archive.contents(of: entry))
                _ = context.chmod(entry.path, mode: FileMode(rawValue: entry.mode))
            }
            for (path, contents) in files {
                Guest.write(context, path, Array(contents.utf8))
            }
            context.exit(0)
        }
        loop.runUntilIdle()

        terminal.onOutput = { [weak self] in
            guard let self else { return }
            self.screenBytes.append(contentsOf: self.terminal.readForApp(max: 65_535))
        }
        // Ctrl-C interrupts the terminal's foreground job, as a console does.
        terminal.onControlC = { [weak self] in
            guard let self else { return }
            self.kernel.interruptProcessGroup(
                self.terminal.foregroundProcessGroupID, signal: Signal.sigint.rawValue)
        }
        let commands = CommandRegistry.builtins
        GoExecutableLoader.register(in: commands)
        kernel.spawn("sh", Programs.shell(tty: terminal.slave, commands: commands))
        settle()
    }

    /// Everything the terminal has displayed, with CR LF folded to LF.
    var screen: String {
        String(decoding: screenBytes, as: UTF8.self)
            .replacingOccurrences(of: "\r\n", with: "\n")
    }

    /// Runs the system until nothing is ready to run. A long guest computation
    /// yields to the event loop every instruction quantum, so it spans more
    /// steps than one drain's budget; keep draining until the loop is idle.
    /// Logical time does not advance, so a sleeping process stays asleep.
    func settle() {
        while loop.runUntilIdle() == .budgetExceeded {}
    }

    /// Types `text` at the terminal and runs the system until it is idle.
    func type(_ text: String) {
        terminal.writeFromApp(Array(text.utf8))
        settle()
    }

    /// Advances logical time by `seconds`, running whatever falls due, such as
    /// a process waking from `time.Sleep`, and then runs the system to idle.
    func advance(by seconds: Double) {
        let target = loop.now + seconds
        while loop.now < target, loop.advance(by: target - loop.now) == .budgetExceeded {}
        settle()
    }

    /// Runs one pipeline with its standard output and standard error captured.
    func run(_ command: String) -> CommandResult {
        type("\(command) > /.stdout 2> /.stderr\n")
        type("echo $? > /.status\n")
        let status = read("/.status").trimmingCharacters(in: .whitespacesAndNewlines)
        return CommandResult(
            output: read("/.stdout"),
            errors: read("/.stderr"),
            status: Int32(status) ?? -1)
    }

    /// Runs one pipeline and returns the exact bytes of its standard output,
    /// or nil when it wrote to standard error or failed.
    func runBytes(_ command: String) -> [UInt8]? {
        type("\(command) > /.stdout 2> /.stderr\n")
        type("echo $? > /.status\n")
        guard read("/.status") == "0\n", read("/.stderr").isEmpty else { return nil }
        return readBytes("/.stdout")
    }

    /// The contents of a guest file as text, or the empty string when it is
    /// missing.
    func read(_ path: String) -> String {
        String(decoding: readBytes(path), as: UTF8.self)
    }

    /// The exact contents of a guest file, or nothing when it is missing.
    func readBytes(_ path: String) -> [UInt8] {
        var bytes: [UInt8] = []
        kernel.spawn("read-file") { context in
            if let descriptor = try? context.openFile(path, access: .readOnly) {
                while let chunk = try? context.readFile(descriptor, max: 65_536),
                    !chunk.isEmpty
                {
                    bytes.append(contentsOf: chunk)
                }
                try? context.closeFile(descriptor)
            }
            context.exit(0)
        }
        settle()
        return bytes
    }

    /// Creates or replaces a guest file.
    func write(_ path: String, _ contents: String) {
        write(path, Array(contents.utf8))
    }

    /// Creates or replaces a guest file with exact bytes.
    func write(_ path: String, _ contents: [UInt8]) {
        kernel.spawn("write-file") { context in
            Guest.write(context, path, contents)
            context.exit(0)
        }
        settle()
    }

    /// Adds `contents` to the end of a guest file, as a logger would.
    func append(_ path: String, _ contents: String) {
        write(path, readBytes(path) + Array(contents.utf8))
    }

    private static func write(_ context: ProcessContext, _ path: String, _ bytes: [UInt8]) {
        let descriptor = context.open(path, create: true, truncate: true)!
        context.write(descriptor, bytes)
        context.close(descriptor)
    }
}

/// One failed check, recorded on the guest's thread and reported by `withGuest`.
struct GuestFailure: Sendable {
    let message: String
    let sourceLocation: SourceLocation
}

extension Guest {
    /// Records a failure unless `condition` holds.
    func check(
        _ condition: Bool,
        _ message: @autoclosure () -> String,
        sourceLocation: SourceLocation = #_sourceLocation
    ) {
        if !condition {
            failures.append(GuestFailure(message: message(), sourceLocation: sourceLocation))
        }
    }
}

/// Runs `body` against a fresh guest and reports every failed check.
///
/// The interpreter's stack frames are large in debug builds, more than the
/// 512 KiB stack of a test task comfortably holds, so the guest lives on its
/// own thread with a generous stack. Swift Testing
/// attributes issues through task-local state that such a thread lacks, so the
/// failures are collected there and recorded here.
func withGuest(
    files: [String: String] = [:],
    _ body: @escaping @Sendable (Guest) throws -> Void
) throws {
    final class Outcome: @unchecked Sendable {
        var failures: [GuestFailure] = []
        var error: (any Error)?
    }
    let outcome = Outcome()
    let finished = DispatchSemaphore(value: 0)
    let thread = Thread {
        do {
            let guest = try Guest(files: files)
            try body(guest)
            outcome.failures = guest.failures
        } catch {
            outcome.error = error
        }
        finished.signal()
    }
    thread.stackSize = 256 << 20
    thread.start()
    finished.wait()
    for failure in outcome.failures {
        Issue.record(Comment(rawValue: failure.message), sourceLocation: failure.sourceLocation)
    }
    if let error = outcome.error { throw error }
}

/// Checks that `command` writes exactly `output` and exits with `status`.
func expect(
    _ guest: Guest,
    _ command: String,
    _ output: String,
    status: Int32 = 0,
    errors: String = "",
    sourceLocation: SourceLocation = #_sourceLocation
) {
    let result = guest.run(command)
    let expected = CommandResult(output: output, errors: errors, status: status)
    guest.check(
        result == expected,
        "\(command)\n  expected \(expected)\n  actual   \(result)",
        sourceLocation: sourceLocation)
}

/// Checks the contents of a guest file.
func expectFile(
    _ guest: Guest,
    _ path: String,
    _ contents: String,
    sourceLocation: SourceLocation = #_sourceLocation
) {
    let actual = guest.read(path)
    guest.check(
        actual == contents,
        "\(path)\n  expected \(contents.debugDescription)\n  actual   \(actual.debugDescription)",
        sourceLocation: sourceLocation)
}

/// Text fixtures shared by the command suites.
enum Fixture {
    static let files: [String: String] = [
        "/fruits": "banana\napple\ncherry\napple\n",
        "/numbers": "10\n9\n-3\n2.5\n2.25\nabc\n0\n-10\n",
        "/pairs": "banana 3\napple 10\nCherry 2\napple 2\n",
        "/table": "a:3:x\nb:1:y\nc:2:x\nd:1:x\n",
        "/blanks": "one\n\n\n\ntwo\tx\n\nthree",
        "/left": "a\nc\ne\n",
        "/right": "b\nc\nd\n",
        "/repeats": "a\nb\nb\nB\nc\nc\nc\n",
        "/utf8": "héllo wörld\n\tfoo  bar\n",
        "/ten": (1...12).map { "line \($0)\n" }.joined(),
    ]
}

/// Checks that `command` succeeds silently and writes exactly `bytes`.
func expectBytes(
    _ guest: Guest,
    _ command: String,
    _ bytes: [UInt8],
    sourceLocation: SourceLocation = #_sourceLocation
) {
    let actual = guest.runBytes(command)
    guest.check(
        actual == bytes,
        "\(command)\n  expected \(bytes.count) bytes \(bytes.prefix(24))…\n  actual   "
            + (actual.map { "\($0.count) bytes \($0.prefix(24))…" } ?? "a failure"),
        sourceLocation: sourceLocation)
}
