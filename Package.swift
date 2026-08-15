// swift-tools-version: 6.3

import PackageDescription

let package = Package(
    name: "SwiftixCoreutils",
    platforms: [.macOS(.v14)],
    dependencies: [
        .package(path: "../Swiftix"),
    ],
    targets: [
        .executableTarget(
            name: "CoreutilsPackageBuilder",
            dependencies: [
                .product(name: "SwiftixGo", package: "Swiftix"),
                .product(name: "SwiftixPackages", package: "Swiftix"),
            ]
        ),
        .testTarget(
            name: "CoreutilsTests",
            dependencies: [
                .product(name: "Swiftix", package: "Swiftix"),
                .product(name: "SwiftixGo", package: "Swiftix"),
                .product(name: "SwiftixPackages", package: "Swiftix"),
            ]
        ),
    ],
    swiftLanguageModes: [.v6]
)
