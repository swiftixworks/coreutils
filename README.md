# Swiftix coreutils

This repository provides the base userland commands for Swiftix. It owns the command sources and builds them into the native `coreutils_1.1.0.pkg` archive. [SwiftixDistribution](https://github.com/swiftixworks/SwiftixDistribution) consumes the validated package and does not duplicate or maintain the command sources.

The package contains 17 commands: `cat`, `comm`, `echo`, `false`, `fold`, `head`, `nl`, `paste`, `rev`, `seq`, `sort`, `tac`, `tail`, `tr`, `true`, `uniq`, and `wc`. They are installed in `/usr/bin`, which the shell searches before its built-in commands, so these are the programs a user runs.

`rev` remains as an extension for compatibility with existing Swiftix Minimal images; Debian provides the equivalent command through `util-linux`.

## Commands

Every command except `echo`, `true`, and `false` parses options the way GNU `getopt` does: short options combine (`-rn`), values attach (`-n5`, `-k2`, `--lines=5`), options may follow operands, `--` ends option parsing, `-` names standard input, and `--help` prints usage on standard output and exits 0. An unknown option prints `CMD: invalid option -- 'x'` and a pointer to `--help` on standard error and exits 1 (`sort`: 2).

| Command | Options | Notes |
| --- | --- | --- |
| `cat` | `-n -b -s -E -T -v -A -e -t -u` | Several files; numbering continues across files. `-v` uses `^X` notation for control characters. Without options the copy is byte for byte and streams. |
| `comm` | `-1 -2 -3 --output-delimiter=STR --total` | Either file may be `-`. |
| `echo` | `-n -e -E` | `-e` expands `\a \b \c \e \f \n \r \t \v \\ \0NNN \xHH`. |
| `fold` | `-w WIDTH`, `-WIDTH`, `-s`, `-b` | Counts characters; `-b` counts bytes. |
| `head` | `-n [-]N`, `-N`, `-c [-]N`, `-q`, `-v` | `==> NAME <==` headers for several files. `N` takes `b`, `K`, `kB`, `M`, `MB` suffixes. `-c` counts exact bytes. Standard input is read only as far as needed. |
| `nl` | `-b a\|t\|n`, `-w N`, `-s STR`, `-v N`, `-i N`, `-n ln\|rn\|rz` | |
| `paste` | `-d LIST`, `-s` | `LIST` cycles and accepts `\n \t \\ \0`. Repeated `-` operands share standard input (`paste - -`). |
| `rev` | | |
| `seq` | `-s STR`, `-w` | `LAST`, `FIRST LAST`, `FIRST INCREMENT LAST`; negative and decimal values. |
| `sort` | `-b -f -h -n -r -u -s -c -C -m`, `-k POS1[,POS2]`, `-t SEP`, `-o FILE` | Stable; several files. Key positions are `F[.C][bfhnr]`. |
| `tac` | `-s STR` | |
| `tail` | `-n [+]N`, `-N`, `-c [+]N`, `-q`, `-v`, `-f`, `-F`, `-s SECONDS` | `-c` counts exact bytes. `-f` follows named files; see Compatibility. |
| `tr` | `-c -C -d -s -t` | Ranges, `\NNN` and the usual escapes, `[:alnum:]` … `[:xdigit:]`, `[c*]`, `[c*n]`. |
| `uniq` | `-c -d -u -i -f N -s N -w N` | Optional `INPUT` and `OUTPUT` operands. |
| `wc` | `-l -w -c -m -L` | Several files with a `total` line; a single count prints bare. |

Long option names follow GNU (`--lines`, `--bytes`, `--reverse`, `--numeric-sort`, `--key`, `--field-separator`, `--unique`, `--count`, and so on); `CMD --help` lists them.

## Compatibility

The commands follow GNU coreutils for the options above in the C locale. They do not claim complete GNU or POSIX compatibility.

- **Diagnostics go to standard error.** Usage errors, the `Try 'CMD --help'` hint, unreadable files (`CMD: PATH: No such file`), and messages such as `sort: /f:2: disorder: …` leave standard output clean, so a failing stage does not feed its message to the next one. The exit status is conventional.
- **Text is UTF-8; three paths are byte-exact.** `cat` without options, `head -c`, and `tail -c` read and write bytes: a count may fall inside a multi-byte character, and binary input passes through unchanged. Every other mode works on text, and a Swiftix Go string is always valid UTF-8, so invalid input bytes are replaced with U+FFFD when read and byte counts (`wc -c`) refer to the repaired text. `fold -b` counts bytes but never splits a character. `echo -e '\xHH'`, `\NNN` in `tr`, and `tr` ranges address the code points 0 to 255.
- **`tail -f` follows named files by polling.** After printing the tail it looks at each file again every second (`-s SECONDS` changes the interval; fractions are accepted) and prints what was appended, with a `==> NAME <==` header when the output switches files. A file that shrank is reported as `tail: NAME: file truncated` on standard error and printed again from its start. The process sleeps between looks, so other processes run meanwhile; it ends when interrupted or when its output is closed. `-F` is the same as `-f`. Standard input is read to its end and not followed further.
- **Input is held in memory.** Apart from plain `cat` and `head` on standard input, a command reads its whole input before writing, up to the runtime's 16 MiB per process. Output of any size is written through a pipe without loss: a writer waits while the pipe is full.
- `sort` compares bytes, cannot mix ascending and descending keys in one invocation, and does not implement `-d`, `-g`, `-i`, `-M`, `-R`, `-V`, or `-z`.
- `cat -v` does not use `M-` notation for non-ASCII characters. `nl` numbers one logical page and has no section delimiters. `tr` has no `[=c=]` equivalence classes.

The commands run on the Swiftix Go virtual machine, which has no limit on how long a run may take but executes guest code far more slowly than native code. `sort` orders its lines with the native `sort.Strings`, and whole-buffer work such as `wc -lwc`, `tr`, and splitting or joining lines is native too, so those cost little at any size. A step that visits every line in Go code takes on the order of ten seconds per 100,000 lines in a debug build (`nl`, `uniq`, `cat -n`, `sort -n`, `seq`), and one that calls a helper function per line (`sort -k`, `rev`, `uniq -f`) takes several times longer on large inputs, because every call allocates its locals on the garbage-collected heap.

Other binary-safe commands such as `base64`, `cksum`, `od`, and `sha256sum` are not part of this package. Commands that need kernel interfaces (`ls`, `cp`, `date`, and similar) are Swiftix built-ins, as are `cut`, `tee`, `basename`, and `dirname`.

## Sources

```text
Commands/<name>/main.go   one command: `package main`, import path `swiftix/Commands/<name>`
internal/cli/cli.go       package `cli` (`swiftix/internal/cli`): option parser, input readers,
                          diagnostics, and string helpers shared by the commands
userland/userland.go      host stand-in for the runtime package `swiftix/userland`
go.mod                    module `swiftix`
package.json              package name, version, and command list
```

The repository is one Go module. Each command is its own `package main` and imports the shared code as an ordinary package:

```go
import "swiftix/internal/cli"

data, status := cli.ReadAll("cat", operands)
lines := cli.SplitLines(data)
```

`cli` lives under `internal/` because it is an implementation detail of these commands, not an API for other modules: Go, and Swiftix Go, refuse to import it from outside this module. Its exported identifiers are the ones commands call; everything else is unexported. `true` and `false` need none of it and import nothing from the module.

The package builder compiles each command as a module build: the command's package plus every local package it imports, linked by the Swiftix Go compiler into one executable image. A call into `cli` costs exactly what a call inside `package main` costs.

`swiftix/userland` is provided by the Swiftix Go runtime (terminal and input primitives such as `ReadStdin` and `ReadInput`). Swiftix Go never compiles the `userland/` directory; it exists so that a standard Go toolchain can resolve the same import path.

### Editing with standard Go tools

The sources are valid standard Go as well as valid Swiftix Go, so gopls, `go build`, `go vet`, and `gofmt` work on the repository as it is. From the repository root, with Go 1.24 or later installed, these report nothing:

```sh
go build -gcflags=-e ./...
go vet ./...
gofmt -l .
```

A host build is for checking only: the `userland` stand-in does not reproduce the Swiftix runtime, and the commands are tested on the Swiftix Go virtual machine.

The sources stay within the subset both compilers accept (see `docs/go-toolchain.md` in the Swiftix repository). The constraints that shape this code:

- `int` and `byte` are the only numeric types, and there are no rune literals, so characters are written as code points (`arg[0] != 45`) and `int(s[i])` converts a string byte for integer arithmetic.
- Each import is its own `import "path"` line; a parenthesized import group is not accepted.
- `time.Sleep` is only ever given a constant, because Swiftix Go has no `time.Duration` type to convert a variable to.
- Function types cannot be written, so helpers take and return plain values rather than callbacks.

## Building and verification

Building requires Swift 6.3 or later and the following sibling repository layout:

```text
swiftixworks/
├── Swiftix/
└── coreutils/
```

Run the following commands from the coreutils repository root:

```sh
swift run CoreutilsPackageBuilder
swift run CoreutilsPackageBuilder --check
swift test -Xswiftc -warnings-as-errors
```

The builder uses the Swiftix Go compiler to produce `svm64` executable images, then assembles them into a deterministic `.pkg` archive with the `SwiftixPackages` codec. Each command is installed at `/usr/bin/<command>`.

The archive is written to `Artifacts/<name>_<version>.pkg`, with both values taken from `package.json`. The `--check` option rebuilds the package and compares it byte for byte with that file. Use `--output <path>` to write an archive to another location.

The tests install the checked-in archive into a fresh Swiftix kernel and run every command through the shell, so rebuild the archive before testing a source change. The `scale` suite pushes 100,000-line inputs through the guest virtual machine and accounts for most of the run time; it prints the wall time of each case.

## License

Swiftix coreutils is available under the [MIT License](LICENSE).
