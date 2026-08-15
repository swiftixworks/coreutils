# Swiftix coreutils

This repository provides the base userland commands for Swiftix. It owns the command sources and builds them into the native `coreutils_1.0.0.pkg` archive. [SwiftixDistribution](https://github.com/swiftixworks/SwiftixDistribution) consumes the validated package and does not duplicate or maintain the command sources.

The package currently contains 17 commands: `cat`, `comm`, `echo`, `false`, `fold`, `head`, `nl`, `paste`, `rev`, `seq`, `sort`, `tac`, `tail`, `tr`, `true`, `uniq`, and `wc`.

The `comm`, `fold`, `paste`, `tac`, and `tr` commands were the first text utilities identified by the distribution plan. `rev` remains as an extension for compatibility with existing Swiftix Minimal images; Debian provides the equivalent command through `util-linux`.

## Compatibility

These implementations target the text-processing capabilities currently available in Swiftix Go. They do not claim complete GNU or POSIX compatibility.

- `tr` supports ASCII ranges and literal UTF-8 characters.
- `fold` counts UTF-8 characters rather than terminal display columns.
- `paste` treats the argument to `-d` as one literal delimiter.

Binary-safe commands such as `base64` and `sha256sum` require runtime support for byte-oriented I/O and integer and bitwise operations. `date` requires an API that exposes the current logical time. These commands remain planned work and are not included in the current package.

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

The `--check` option rebuilds the package and compares it byte for byte with `Artifacts/coreutils_1.0.0.pkg`. Use `--output <path>` to write an archive to another location.

## License

Swiftix coreutils is available under the [MIT License](LICENSE).
