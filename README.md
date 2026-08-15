# Swiftix coreutils

Swiftix 的基础用户空间命令。这个仓库拥有命令源码，并把它们构建为原生
`coreutils_1.0.0.pkg`；SwiftixDistribution 只消费已验证的软件包，不再复制或维护命令源码。

当前包含 17 个命令：`cat`、`comm`、`echo`、`false`、`fold`、`head`、`nl`、
`paste`、`rev`、`seq`、`sort`、`tac`、`tail`、`tr`、`true`、`uniq`、`wc`。
其中 `comm`、`fold`、`paste`、`tac`、`tr` 是此前发行版计划中的首批文本工具；
`rev` 是为兼容既有 Swiftix Minimal 保留的扩展（Debian 将它放在 `util-linux`）。

这些实现面向当前 Swiftix Go 的文本能力，不宣称完整兼容 GNU/POSIX：`tr` 支持
ASCII 范围和 UTF-8 字面字符，`fold` 按 UTF-8 字符数而非终端显示宽度折行，`paste`
把 `-d` 参数作为一个字面分隔符。二进制安全的 `base64`、`sha256sum` 等继续等待
运行时的 byte I/O 与整数/位运算能力，`date` 则等待可读取当前逻辑时间的 API；
它们仍属于 coreutils 后续规划，不冒充本版已经交付。

## 构建与验证

本仓库与 Swiftix 采用同级目录布局：

```text
swiftixworks/
├── Swiftix/
└── coreutils/
```

```sh
swift run CoreutilsPackageBuilder
swift run CoreutilsPackageBuilder --check
swift test -Xswiftc -warnings-as-errors
```

构建器使用 Swiftix Go 编译器生成 `svm64` executable image，再通过
`SwiftixPackages` codec 组装确定性的 `.pkg`。归档安装路径固定为 `/usr/bin/<command>`。

## 许可证

本项目采用 [MIT License](LICENSE)。
