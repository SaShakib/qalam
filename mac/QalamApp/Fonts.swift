import AppKit
import CoreText

/// Arabic fonts: four open-licence fonts ship inside the app, and KFGQPC Hafs when it's in fonts/ at build
/// time; an installed KFGQPC Hafs is found too.
enum ArabicFonts {
    struct Choice: Identifiable {
        var id: String { family }
        let family: String
        let note: String
    }

    static let bundled: [Choice] = [
        Choice(family: "Scheherazade New", note: "SIL, large and very readable harakat"),
        Choice(family: "Noto Naskh Arabic", note: "clear naskh"),
        Choice(family: "Amiri Quran", note: "classic mushaf style"),
        Choice(family: "Noto Sans Arabic", note: "modern sans-serif"),
    ]

    /// The default Arabic font (SIL Scheherazade New, included with Qalam).
    static let defaultArabic = "Scheherazade New"

    /// The KFGQPC Hafs family (the Madinah mushaf font) if it's usable: shipped with Qalam or installed.
    static func kfgqpcAvailable() -> String? {
        let names = (CTFontManagerCopyAvailableFontFamilyNames() as? [String]) ?? []
        return names.first { $0.uppercased().contains("KFGQPC") && $0.uppercased().contains("HAFS") }
            ?? names.first { $0.uppercased().contains("KFGQPC") }
    }

    /// The default Qur'an font: KFGQPC Hafs, the standard mushaf font, or Scheherazade New until it's there.
    static func defaultQuran() -> String { kfgqpcAvailable() ?? defaultArabic }

    // MARK: Fonts for every app

    /// The fonts worth having in every app (Pages, Word, Notes…): SIL Scheherazade New, and KFGQPC Hafs
    /// when it's shipped with Qalam. Other apps can't see the fonts that are only inside Qalam.
    static var forAllApps: [URL] {
        guard let dir = Bundle.main.resourceURL?.appendingPathComponent("Fonts"),
              let files = try? FileManager.default.contentsOfDirectory(at: dir, includingPropertiesForKeys: nil) else { return [] }
        return files.filter {
            let name = $0.lastPathComponent.uppercased()
            return $0.pathExtension.lowercased() == "ttf" && (name.hasPrefix("SCHEHERAZADE") || name.contains("KFGQPC") || name.contains("HAFS"))
        }.sorted { $0.lastPathComponent < $1.lastPathComponent }
    }

    static var userFontsFolder: URL {
        FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Library/Fonts")
    }

    /// True when every font in `forAllApps` is in ~/Library/Fonts.
    static var installedForAllApps: Bool {
        forAllApps.allSatisfy { FileManager.default.fileExists(atPath: userFontsFolder.appendingPathComponent($0.lastPathComponent).path) }
    }

    /// Copies the fonts to ~/Library/Fonts (no password needed); macOS makes them available to every app.
    @discardableResult
    static func installForAllApps() -> Bool {
        let fm = FileManager.default
        try? fm.createDirectory(at: userFontsFolder, withIntermediateDirectories: true)
        for f in forAllApps {
            let dest = userFontsFolder.appendingPathComponent(f.lastPathComponent)
            if !fm.fileExists(atPath: dest.path) { try? fm.copyItem(at: f, to: dest) }
        }
        return installedForAllApps
    }

    static let kfgqpcDownload = URL(string: "https://fonts.qurancomplex.gov.sa/")!

    /// Arabic-capable fonts on this Mac, and the KFGQPC Hafs family if present.
    /// Uses CoreText only, so it is safe to run off the main thread.
    static func scanInstalled() -> ([String], String?) {
        let names = (CTFontManagerCopyAvailableFontFamilyNames() as? [String]) ?? []
        let skip = Set(bundled.map(\.family))
        var arabic: [String] = []
        for family in names where !skip.contains(family) && !family.hasPrefix(".") {
            let font = CTFontCreateWithName(family as CFString, 12, nil)
            let set = CTFontCopyCharacterSet(font) as CharacterSet
            let actual = CTFontCopyFamilyName(font) as String
            if actual == family, set.contains(Unicode.Scalar(0x0628)!) { arabic.append(family) }
        }
        let kfgqpc = arabic.first { $0.uppercased().contains("KFGQPC") && $0.uppercased().contains("HAFS") }
            ?? arabic.first { $0.uppercased().contains("KFGQPC") }
        return (arabic.sorted(), kfgqpc)
    }

    /// Makes the bundled fonts available to this app (no system install needed).
    static func registerBundled() {
        var dirs: [URL] = []
        if let res = Bundle.main.resourceURL { dirs.append(res.appendingPathComponent("Fonts")) }
        dirs.append(URL(fileURLWithPath: FileManager.default.currentDirectoryPath).appendingPathComponent("fonts"))
        for dir in dirs {
            guard let files = try? FileManager.default.contentsOfDirectory(at: dir, includingPropertiesForKeys: nil) else { continue }
            for f in files where f.pathExtension.lowercased() == "ttf" {
                CTFontManagerRegisterFontsForURL(f as CFURL, .process, nil)
            }
            return
        }
    }
}
