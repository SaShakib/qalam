import Foundation
import QalamEngine

// qalam ayah 2:255-257 [--trans en.sahih|bn.bengali|bn.hoque|all|none] [--insert]
// qalam search رحمة | rahma | "mercy" [--trans …] [--max 50]
// qalam surahs | qalam translations
// Data: $QALAM_QURAN, quran/ next to the repo, then the installed Qalam.app.
enum QuranCLI {
    static func library() -> QuranLibrary? {
        var dirs: [URL] = []
        if let env = ProcessInfo.processInfo.environment["QALAM_QURAN"] { dirs.append(URL(fileURLWithPath: env)) }
        let exe = URL(fileURLWithPath: CommandLine.arguments[0]).resolvingSymlinksInPath()
        var d = exe.deletingLastPathComponent()
        for _ in 0..<5 { dirs.append(d.appendingPathComponent("quran")); d.deleteLastPathComponent() }   // .build/release/qalam → repo
        dirs.append(URL(fileURLWithPath: FileManager.default.currentDirectoryPath).appendingPathComponent("quran"))
        let home = FileManager.default.homeDirectoryForCurrentUser
        dirs += [home.appendingPathComponent("Applications/Qalam.app/Contents/Resources/Quran"),
                 URL(fileURLWithPath: "/Applications/Qalam.app/Contents/Resources/Quran")]
        return QuranLibrary(folders: dirs)
    }

    static func run(_ args: [String]) -> Int32 {
        guard let lib = library() else { print("Qur'an data not found (set QALAM_QURAN to the quran/ folder)"); return 2 }
        var words: [String] = []
        var trans: [String] = ["en.sahih"]
        var insert = false
        var maxHits = 50
        var i = 1
        while i < args.count {
            switch args[i] {
            case "--trans" where i + 1 < args.count:
                i += 1
                let t = args[i]
                trans = t == "none" ? [] : t == "all" ? lib.available().map(\.id) : t.split(separator: ",").map(String.init)
            case "--insert": insert = true
            case "--max" where i + 1 < args.count: i += 1; maxHits = Int(args[i]) ?? maxHits
            default: words.append(args[i])
            }
            i += 1
        }
        for t in trans where lib.translation(t) == nil || !lib.available().contains(where: { $0.id == t }) {
            print("translation \(t) is not on this Mac; have: \(lib.available().map(\.id).joined(separator: ", "))"); return 2
        }
        let query = words.joined(separator: " ")
        switch args[0] {
        case "surahs":
            for s in lib.surahs { print("\(s.number)\t\(s.ayahs)\t\(s.arabic)\t\(s.name)\t\(s.english)\t\(s.type)") }
        case "translations":
            for t in lib.catalog { print("\(t.id)\t\(lib.available().contains(t) ? "here" : "download in the app")\t\(t.title)") }
        case "ayah":
            // several references allowed: qalam ayah 2:255 3:18-19
            for ref in words {
                guard let p = lib.resolve(ref) else { print("? \(ref): not a reference (try 2:255 or baqarah 255-257)"); return 1 }
                if insert { print(lib.text(p, trans.isEmpty ? .arabic : .both, translation: trans.first)); continue }
                for a in p.ayahs { show(lib, p.surah, a, trans) }
            }
        case "search":
            let typedArabic = query.unicodeScalars.contains { (0x0600...0x06FF).contains($0.value) }
            let ar = typedArabic ? query : Qalam.text(query, Options(harakat: .none))
            let (hits, total) = lib.searchArabic(ar, phonetic: !typedArabic, limit: maxHits)
            if total > 0 || typedArabic { print("# Arabic \(ar): \(total) āyāt\(total > hits.count ? " (first \(hits.count))" : "")") }
            for p in hits { show(lib, p.surah, p.from, trans) }
            for t in trans where !query.unicodeScalars.contains(where: { (0x0600...0x06FF).contains($0.value) }) {
                let (th, tt) = lib.searchTranslation(t, query, limit: maxHits)
                print("# \(t) “\(query)”: \(tt) āyāt\(tt > th.count ? " (first \(th.count))" : "")")
                for p in th { show(lib, p.surah, p.from, [t]) }
            }
        default: return 2
        }
        return 0
    }

    static func show(_ lib: QuranLibrary, _ s: Int, _ a: Int, _ trans: [String]) {
        print("[\(s):\(a)] \(lib.arabic(s, a) ?? "")")
        for t in trans { print("  \(t): \(lib.translation(t, s, a) ?? "")") }
    }
}
