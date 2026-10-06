import AppKit
import CoreText

/// Arabic fonts: four open-licence fonts ship inside the app; KFGQPC Hafs is used if installed.
enum ArabicFonts {
    struct Choice: Identifiable {
        var id: String { family }
        let family: String
        let note: String
    }

    static let bundled: [Choice] = [
        Choice(family: "Noto Naskh Arabic", note: "clear naskh, best for everyday text"),
        Choice(family: "Amiri Quran", note: "classic mushaf style, best for Qur'an"),
        Choice(family: "Scheherazade New", note: "SIL, large and very readable harakat"),
        Choice(family: "Noto Sans Arabic", note: "modern sans-serif"),
    ]

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
