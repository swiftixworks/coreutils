import Foundation
import Swiftix
import SwiftixGoRuntime
import SwiftixPackages
import Testing

@Suite("coreutils package")
struct CoreutilsPackageTests {
    @Test("checked-in archive is a native svm64 package with every command")
    func archiveContents() throws {
        let archive = try loadArchive()

        #expect(archive.manifest.name == "coreutils")
        #expect(archive.manifest.version.description == "1.0.0")
        #expect(archive.manifest.architecture == "svm64")
        let commandPaths = Set(commandNames.map { "/usr/bin/" + $0 })
        #expect(Set(archive.files.map(\.path)) == commandPaths.union([
            "/usr/share/doc/coreutils/LICENSE"
        ]))
        for entry in archive.files where commandPaths.contains(entry.path) {
            #expect(entry.mode == 0o755)
            #expect(GoExecutableImage.recognizes(archive.contents(of: entry)))
        }
        #expect(archive.files.first { $0.path.hasSuffix("/LICENSE") }?.mode == 0o644)
    }

    @Test("planned text commands execute from the packaged images")
    func plannedTextCommandsExecute() throws {
        let archive = try loadArchive()
        let loop = EventLoop()
        let kernel = Kernel(loop: loop)
        kernel.spawn("install-coreutils") { context in
            _ = context.mkdir("/usr")
            _ = context.mkdir("/usr/bin")
            for entry in archive.files where entry.path.hasPrefix("/usr/bin/") {
                let descriptor = context.open(entry.path, create: true, truncate: true)!
                context.write(descriptor, archive.contents(of: entry))
                context.close(descriptor)
                _ = context.chmod(entry.path, mode: FileMode(rawValue: entry.mode))
            }
            for (path, contents) in [
                ("/left", "a\nshared\n"),
                ("/right", "b\nshared\n"),
            ] {
                let descriptor = context.open(path, create: true, truncate: true)!
                context.write(descriptor, Array(contents.utf8))
                context.close(descriptor)
            }
            context.exit(0)
        }
        loop.runUntilIdle()

        let terminal = PseudoTerminal()
        var output: [UInt8] = []
        terminal.onOutput = { [weak terminal] in
            guard let terminal else { return }
            output.append(contentsOf: terminal.readForApp(max: 65_535))
        }
        let commands = CommandRegistry.builtins
        GoExecutableLoader.register(in: commands)
        kernel.spawn("sh", Programs.shell(tty: terminal.slave, commands: commands))
        loop.runUntilIdle()
        for line in [
            "echo banana | tr a-z A-Z",
            "echo aaabbb | tr -s ab",
            "echo abc | tr -d b",
            "tac /left",
            "paste -d , /left /right",
            "comm /left /right",
            "comm -1 -2 /left /right",
            "echo abcdef | fold -w 3",
        ] {
            terminal.writeFromApp(Array((line + "\n").utf8))
            loop.runUntilIdle()
        }

        let rendered = String(decoding: output, as: UTF8.self)
        #expect(rendered.contains("BANANA\n"))
        #expect(rendered.contains("ab\n"))
        #expect(rendered.contains("ac\n"))
        #expect(rendered.contains("shared\na\n"))
        #expect(rendered.contains("a,b\nshared,shared\n"))
        #expect(rendered.contains("a\n\tb\n\t\tshared\n"))
        #expect(rendered.contains("comm -1 -2 /left /right\nshared\n"))
        #expect(rendered.contains("abc\ndef\n"))
    }

    private var commandNames: [String] {
        [
            "cat", "comm", "echo", "false", "fold", "head", "nl", "paste",
            "rev", "seq", "sort", "tac", "tail", "tr", "true", "uniq", "wc",
        ]
    }

    private func loadArchive() throws -> PackageArchive {
        let root = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .deletingLastPathComponent()
        let bytes = Array(try Data(contentsOf: root
            .appendingPathComponent("Artifacts/coreutils_1.0.0.pkg")))
        return try PackageArchive.decode(bytes)
    }
}
