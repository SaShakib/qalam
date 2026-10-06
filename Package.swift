// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "Qalam",
    platforms: [.macOS(.v13)],
    targets: [
        // Pure transliteration engine: no UI, shared by everything below.
        .target(name: "QalamEngine", path: "engine"),
        // `qalam kataba` → كَتَبَ, and `qalam --test tests/cases.tsv`.
        .executableTarget(name: "qalam", dependencies: ["QalamEngine"], path: "cli"),
        // The input method (lives in ~/Library/Input Methods).
        .executableTarget(
            name: "QalamInput",
            dependencies: ["QalamEngine"],
            path: "mac/QalamInput",
            linkerSettings: [.linkedFramework("InputMethodKit")]
        ),
        // The guide + practice app (lives in ~/Applications).
        .executableTarget(
            name: "QalamApp",
            dependencies: ["QalamEngine"],
            path: "mac/QalamApp",
            linkerSettings: [.linkedFramework("Carbon")]
        ),
    ]
)
