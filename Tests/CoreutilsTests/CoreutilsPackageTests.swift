import Foundation
import SwiftixGoRuntime
import SwiftixPackages
import Testing

@Suite("coreutils package")
struct CoreutilsPackageTests {
    @Test("checked-in archive is a native svm64 package with every command")
    func archiveContents() throws {
        let archive = try CoreutilsArchive.load()

        #expect(archive.manifest.name == "coreutils")
        #expect(archive.manifest.version.description == CoreutilsArchive.version)
        #expect(archive.manifest.architecture == "svm64")
        let commandPaths = Set(CoreutilsArchive.commandNames.map { "/usr/bin/" + $0 })
        #expect(Set(archive.files.map(\.path)) == commandPaths.union([
            "/usr/share/doc/coreutils/LICENSE"
        ]))
        for entry in archive.files where commandPaths.contains(entry.path) {
            #expect(entry.mode == 0o755)
            #expect(GoExecutableImage.recognizes(archive.contents(of: entry)))
        }
        #expect(archive.files.first { $0.path.hasSuffix("/LICENSE") }?.mode == 0o644)
    }

    @Test("package.json names the archive version and the packaged commands")
    func specificationMatchesArchive() throws {
        struct Specification: Decodable {
            let version: String
            let commands: [String]
        }
        let root = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .deletingLastPathComponent()
        let specification = try JSONDecoder().decode(
            Specification.self,
            from: Data(contentsOf: root.appendingPathComponent("package.json")))
        #expect(specification.version == CoreutilsArchive.version)
        #expect(specification.commands == CoreutilsArchive.commandNames)
    }

    @Test("packaged commands shadow the shell's built-in fallbacks through PATH")
    func packagedCommandsRunFromPath() throws {
        try withGuest(files: Fixture.files) { guest in
            // The built-in `sort` and `wc` reject these options, so a correct
            // answer proves the packaged images were the ones resolved.
            expect(guest, "sort -rn /numbers | head -n 1", "10\n")
            expect(guest, "wc -l /fruits", "4 /fruits\n")
            expect(guest, "/usr/bin/echo -n direct", "direct")
        }
    }
}
