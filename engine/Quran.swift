import Foundation

// The whole Qur'an (Tanzil Uthmani text, full harakat) and its translations, for the keyboard's
// "insert āyah" box (⌃⇧Q), the app's Qur'an page and `qalam ayah 2:255`.
// Data: quran/ in the repo, Qalam.app/Contents/Resources/Quran when installed, and translations the
// app downloads into ~/Library/Application Support/Qalam/Quran. Files are Tanzil "sura|aya|text" lines.

public struct Surah: Identifiable, Hashable, Sendable {
    public let number: Int
    public let ayahs: Int
    public let arabic: String
    public let name: String       // Tanzil transliteration, e.g. "Al-Baqara"
    public let english: String    // "The Cow"
    public let type: String       // Meccan / Medinan
    public var id: Int { number }
}

public struct QuranTranslation: Identifiable, Hashable, Sendable {
    public let id: String         // Tanzil id, e.g. "en.sahih"
    public let language: String
    public let name: String
    public let translator: String
    public let bundled: Bool
    /// "English: Saheeh International", "Bangla: Muhiuddin Khan (মুহিউদ্দীন খান)".
    public var title: String {
        name == translator || name.allSatisfy(\.isASCII) && translator.hasPrefix(name)
            ? "\(language): \(name)" : "\(language): \(translator) (\(name))"
    }
    public var short: String { name.allSatisfy(\.isASCII) ? name : translator }
}

/// What the keyboard puts in: Arabic only, Arabic with a translation, or the translation only.
public enum QuranInsert: String, CaseIterable, Identifiable, Sendable {
    case arabic, both, translation
    public var id: String { rawValue }
    public var title: String {
        switch self {
        case .arabic: return "Arabic only"
        case .both: return "Arabic + translation"
        case .translation: return "Translation only"
        }
    }
}

/// A run of āyāt inside one sūrah.
public struct Passage: Hashable, Sendable {
    public let surah: Int
    public let from: Int
    public let to: Int
    public init(surah: Int, from: Int, to: Int) { self.surah = surah; self.from = from; self.to = to }
    public var ayahs: ClosedRange<Int> { from...to }
    public var reference: String { from == to ? "\(surah):\(from)" : "\(surah):\(from)-\(to)" }
}

/// One row of the dropdown under the query.
public enum QuranSuggestion: Hashable, Sendable {
    case surah(Surah)                 // pick → the query becomes "2:" so the āyah number comes next
    case passage(Passage)             // pick → insert
    case hit(Passage, arabic: Bool)   // a search match (in the Arabic, or in the translation)
    case problem(String)              // e.g. "Al-Baqara has 286 āyāt"
}

public final class QuranLibrary {
    public let surahs: [Surah]
    /// Every translation Qalam knows about (bundled or downloadable), from translations.tsv.
    public let catalog: [QuranTranslation]
    private let folders: [URL]
    private let start: [Int]                     // index of each sūrah's first āyah (1-based sūrah)
    private var texts: [String: [String]] = [:]  // "quran-uthmani" or a translation id → 6236 lines
    private let lock = NSLock()

    public static let total = 6236
    static let basmala = "بِسْمِ ٱللَّهِ ٱلرَّحْمَٰنِ ٱلرَّحِيمِ "

    /// Where downloaded translations go (shared by the app, the keyboard and the command line).
    public static var userFolder: URL {
        FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Library/Application Support/Qalam/Quran")
    }

    /// `folders`: the first one that has quran-uthmani.txt is the main one; the user folder is always added.
    public init?(folders: [URL]) {
        let fm = FileManager.default
        guard let main = folders.first(where: { fm.fileExists(atPath: $0.appendingPathComponent("quran-uthmani.txt").path) }),
              let surahText = try? String(contentsOf: main.appendingPathComponent("surahs.tsv"), encoding: .utf8)
        else { return nil }
        var list: [Surah] = []
        for line in surahText.split(separator: "\n") where !line.hasPrefix("#") {
            let c = line.split(separator: "\t", omittingEmptySubsequences: false).map(String.init)
            guard c.count >= 6, let n = Int(c[0]), let a = Int(c[1]) else { continue }
            list.append(Surah(number: n, ayahs: a, arabic: c[2], name: c[3], english: c[4], type: c[5]))
        }
        guard list.count == 114 else { return nil }
        surahs = list
        var s = [0, 0]
        for x in list.dropLast() { s.append(s.last! + x.ayahs) }
        start = s
        var cat: [QuranTranslation] = []
        if let t = try? String(contentsOf: main.appendingPathComponent("translations.tsv"), encoding: .utf8) {
            for line in t.split(separator: "\n") where !line.hasPrefix("#") {
                let c = line.split(separator: "\t").map(String.init)
                if c.count >= 5 { cat.append(QuranTranslation(id: c[0], language: c[1], name: c[2], translator: c[3], bundled: c[4] == "yes")) }
            }
        }
        catalog = cat
        self.folders = [main] + folders.filter { $0 != main } + [Self.userFolder]
    }

    // MARK: Text

    /// Translations whose file is on this Mac, bundled ones first, in catalog order.
    public func available() -> [QuranTranslation] {
        catalog.filter { file(for: $0.id) != nil }
    }

    public func translation(_ id: String) -> QuranTranslation? { catalog.first { $0.id == id } }

    public func surah(_ n: Int) -> Surah? { (1...114).contains(n) ? surahs[n - 1] : nil }

    /// The Arabic of one āyah, with full harakat (the basmala Tanzil puts before āyah 1 is left out).
    public func arabic(_ surah: Int, _ ayah: Int) -> String? {
        guard let t = lines("quran-uthmani"), let i = index(surah, ayah) else { return nil }
        let s = t[i]
        if ayah == 1 && surah != 1 && s.hasPrefix(Self.basmala) { return String(s.dropFirst(Self.basmala.count)) }
        return s
    }

    public func translation(_ id: String, _ surah: Int, _ ayah: Int) -> String? {
        guard let t = lines(id), let i = index(surah, ayah) else { return nil }
        return t[i]
    }

    /// Forget a loaded file (after a download replaced or added it).
    public func reload(_ id: String) {
        lock.lock(); texts[id] = nil; lock.unlock()
    }

    private func index(_ surah: Int, _ ayah: Int) -> Int? {
        guard let s = self.surah(surah), (1...s.ayahs).contains(ayah) else { return nil }
        return start[surah] + ayah - 1
    }

    private func file(for id: String) -> URL? {
        folders.map { $0.appendingPathComponent(id + ".txt") }.first { FileManager.default.fileExists(atPath: $0.path) }
    }

    private func lines(_ id: String) -> [String]? {
        lock.lock(); defer { lock.unlock() }
        if let t = texts[id] { return t }
        guard let url = file(for: id), let t = Self.parse(url) else { return nil }
        texts[id] = t
        return t
    }

    /// "sura|aya|text" lines in order → the 6236 texts; nil if the file isn't a whole Qur'an.
    static func parse(_ url: URL) -> [String]? {
        guard let s = try? String(contentsOf: url, encoding: .utf8) else { return nil }
        return parse(s)
    }

    public static func parse(_ s: String) -> [String]? {
        var out: [String] = []
        out.reserveCapacity(total)
        for line in s.split(separator: "\n", omittingEmptySubsequences: true) {
            guard let c = line.first, c.isNumber else { continue }
            let parts = line.split(separator: "|", maxSplits: 2, omittingEmptySubsequences: false)
            guard parts.count == 3 else { continue }
            out.append(parts[2].trimmingCharacters(in: .whitespacesAndNewlines))
        }
        return out.count == total ? out : nil
    }

    // MARK: Query: "2:255", "2:255-257", "2 255", "baqarah 255", "al imran 3-5", "cow", "2"

    public struct Query: Equatable {
        public var surahs: [Surah]     // what the sūrah part matches (best first)
        public var from: Int?
        public var to: Int?
    }

    public func parse(query raw: String) -> Query {
        var q = raw.trimmingCharacters(in: .whitespaces).replacingOccurrences(of: "–", with: "-")
        var from: Int?, to: Int?
        // trailing "255" or "255-257" (or "255-") is the āyah part, if something comes before it
        if let r = q.range(of: #"[\s:.]\s*(\d+)\s*(-\s*(\d*))?\s*$"#, options: .regularExpression) {
            let tail = String(q[r])
            let nums = tail.split(whereSeparator: { !$0.isNumber }).compactMap { Int($0) }
            from = nums.first
            to = nums.count > 1 ? nums[1] : nil
            q = String(q[..<r.lowerBound])
        }
        q = q.trimmingCharacters(in: CharacterSet(charactersIn: " :.-"))
        return Query(surahs: matchSurahs(q), from: from, to: to)
    }

    /// The dropdown for what's typed so far: a reference ("2:255", "baqarah 255"), sūrah names, and
    /// search results. `arabic` is the query as Arabic (typed in Arabic, or Latin through Qalam).
    public func suggestions(for raw: String, arabic: String? = nil, translation id: String? = nil, limit: Int = 8) -> [QuranSuggestion] {
        let trimmed = raw.trimmingCharacters(in: .whitespaces)
        if trimmed.isEmpty { return [] }
        let q = parse(query: raw)
        guard let first = q.surahs.first, q.from != nil else {
            // Not a reference: sūrah names first, then āyāt that contain the words.
            var rows: [QuranSuggestion] = q.surahs.prefix(q.from == nil ? 4 : 0).map { .surah($0) }
            let typedArabic = trimmed.unicodeScalars.contains { (0x0600...0x06FF).contains($0.value) }
            var inArabic: [QuranSuggestion] = [], inTranslation: [QuranSuggestion] = []
            if let ar = typedArabic ? trimmed : arabic {
                inArabic = searchArabic(ar, phonetic: !typedArabic, limit: limit).hits.map { .hit($0, arabic: true) }
            }
            if let id, !typedArabic, !trimmed.allSatisfy({ $0.isNumber || $0 == ":" }) {
                inTranslation = searchTranslation(id, trimmed, limit: limit).hits.map { .hit($0, arabic: false) }
            }
            // Latin that reads as a word of the translation ("mercy") → those first; otherwise the Arabic.
            rows += inTranslation.isEmpty ? inArabic : inTranslation + inArabic.prefix(2)
            if rows.isEmpty { rows = [.problem("Nothing found for “\(trimmed)”")] }
            return Array(rows.prefix(limit))
        }
        guard let from = q.from else { return [] }
        // A clear name or number with an āyah: that passage (other sūrahs with the same start below it).
        var rows: [QuranSuggestion] = []
        for s in q.surahs.prefix(4) {
            if let p = passage(s, from, q.to) {
                rows.append(.passage(p))
            } else if s == first {
                rows.append(.problem("\(s.name) has \(s.ayahs) āyāt"))
            }
        }
        return rows
    }

    /// The passage a query means, or nil.
    public func resolve(_ raw: String) -> Passage? {
        let q = parse(query: raw)
        guard let s = q.surahs.first else { return nil }
        guard let from = q.from else { return Passage(surah: s.number, from: 1, to: s.ayahs) }   // "112" = whole sūrah
        return passage(s, from, q.to)
    }

    private func passage(_ s: Surah, _ from: Int, _ to: Int?) -> Passage? {
        guard (1...s.ayahs).contains(from) else { return nil }
        let end = min(max(to ?? from, from), s.ayahs)
        return Passage(surah: s.number, from: from, to: end)
    }

    func matchSurahs(_ spec: String) -> [Surah] {
        if spec.isEmpty { return [] }
        if let n = Int(spec) { return surah(n).map { [$0] } ?? [] }
        let key = Self.fold(spec)
        guard !key.isEmpty else { return [] }
        var exact: [Surah] = [], prefix: [Surah] = []
        for s in surahs {
            let keys = keysCache[s.number]
            if keys.contains(key) { exact.append(s) } else if keys.contains(where: { $0.hasPrefix(key) }) { prefix.append(s) }
        }
        return exact + prefix
    }

    private lazy var keysCache: [[String]] = [[]] + surahs.map { s in
        var k = [Self.fold(s.name), Self.fold(s.english), Self.fold(s.arabic)]
        let parts = s.name.split(separator: "-")
        for i in 1..<max(parts.count, 1) { k.append(Self.fold(parts[i...].joined())) }   // "Baqara", "Imraan"
        if s.english.hasPrefix("The ") { k.append(Self.fold(String(s.english.dropFirst(4)))) }
        return Array(Set(k.filter { !$0.isEmpty }))
    }

    /// Loose spelling: "Al-Baqara", "al baqarah", "albaqara" → "albaqara"; "Yaseen" = "yasin"; "Faatiha" = "fatiha".
    static func fold(_ s: String) -> String {
        let arabic = s.unicodeScalars.contains { (0x0600...0x06FF).contains($0.value) }
        if arabic {   // bare letters only: البقرة, بقرة
            var a = String(s.unicodeScalars.filter { (0x0621...0x064A).contains($0.value) }.map(Character.init))
            a = a.replacingOccurrences(of: "أ", with: "ا").replacingOccurrences(of: "إ", with: "ا").replacingOccurrences(of: "آ", with: "ا")
            if a.hasPrefix("ال") { a.removeFirst(2) }
            return a
        }
        var t = s.lowercased().filter { $0.isASCII && $0.isLetter }
        for (a, b) in [("aa", "a"), ("ee", "i"), ("ii", "i"), ("oo", "u"), ("uu", "u"), ("ou", "u"), ("y", "i"), ("w", "u")] {
            t = t.replacingOccurrences(of: a, with: b)
        }
        var out = ""
        for c in t where out.last != c { out.append(c) }   // doubled letters count once
        if out.count > 3 && out.hasSuffix("h") { out.removeLast() }   // baqarah = baqara
        return out
    }

    // MARK: Search: the whole Qur'an, with or without harakat, and the translation

    /// Arabic without harakat or Qur'anic marks, two ways (dagger alif dropped / as ا), so
    /// رحمن and رحمان, صلوة and صلاة both find ٱلرَّحْمَٰنِ / ٱلصَّلَوٰةَ.
    private lazy var bareArabic: [(String, String)] = (lines("quran-uthmani") ?? []).enumerated().map { i, s in
        var t = s
        if let (su, a) = position(i), a == 1, su != 1, t.hasPrefix(Self.basmala) { t = String(t.dropFirst(Self.basmala.count)) }
        return (Self.bare(t, daggerAlif: false), Self.bare(t, daggerAlif: true))
    }
    private lazy var lightArabic: [String] = (lines("quran-uthmani") ?? []).map(Self.light)
    private var foldedTranslation: [String: [String]] = [:]

    /// Letters only: harakat, small Qur'anic marks and tatweel removed; ٱ أ إ آ → ا, ى → ي, ة → ه.
    public static func bare(_ s: String, daggerAlif: Bool = false) -> String {
        var out = String.UnicodeScalarView()
        // Uthmani وٰ / ىٰ is written ا today: ٱلصَّلَوٰةَ = الصلاة, ٱلتَّوْرَىٰةَ = التوراة
        let src = daggerAlif ? s.replacingOccurrences(of: "\u{0648}\u{0670}", with: "ا").replacingOccurrences(of: "\u{0649}\u{0670}", with: "ا") : s
        for u in src.unicodeScalars {
            switch u.value {
            case 0x0670: if daggerAlif { out.append("ا") }
            case 0x064B...0x065F, 0x06D6...0x06ED, 0x0640, 0x08F0...0x08FF, 0x200C, 0x200D: continue
            case 0x0671, 0x0623, 0x0625, 0x0622: out.append("ا")
            case 0x0649: out.append("ي")
            case 0x0629: out.append("ه")
            default: out.append(u)
            }
        }
        return String(out).replacingOccurrences(of: "اا", with: "ا")
    }

    /// For Latin typed phonetically: letters an English typist mixes up count as one (ح = ه, ص = س, ط = ت,
    /// ض = د, ظ ذ = ز, ق = ك, ع ء = ا, ث = س).
    static func loose(_ bare: String) -> String {
        let map: [Character: Character] = ["ح": "ه", "خ": "ه", "ص": "س", "ث": "س", "ط": "ت", "ض": "د", "ظ": "ز", "ذ": "ز",
                                           "ق": "ك", "ع": "ا", "ء": "ا", "ؤ": "و", "ئ": "ي", "غ": "ج"]
        var out = ""
        for c in bare { let m = map[c] ?? c; if !(m == "ا" && out.last == "ا") { out.append(m) } }
        return out
    }
    private lazy var looseArabic: [(String, String)] = bareArabic.map { (Self.loose($0.0), Self.loose($0.1)) }

    /// Harakat kept, only spelling variants evened out (ۡ = ْ, ٱ = ا, no small marks), for ranking typed harakat.
    static func light(_ s: String) -> String {
        var out = String.UnicodeScalarView()
        for u in s.unicodeScalars {
            switch u.value {
            case 0x06E1: out.append("\u{0652}")
            case 0x0671: out.append("ا")
            case 0x06D6...0x06E0, 0x06E2...0x06ED, 0x0640, 0x200C, 0x200D: continue
            default: out.append(u)
            }
        }
        return String(out)
    }

    /// Case, accents, joiners and Bangla nukta forms (য় = য + ়) evened out.
    static func foldText(_ s: String) -> String {
        let scalars = Array(s.decomposedStringWithCanonicalMapping.lowercased().unicodeScalars)
        var out = String.UnicodeScalarView()
        for (i, u) in scalars.enumerated() {
            switch u.value {
            case 0x0300...0x036F, 0x200C, 0x200D: continue                      // accents, joiners
            case 0x09CD where i + 1 == scalars.count || !(0x0995...0x09B9).contains(scalars[i + 1].value):
                continue                                                         // আল্লাহ্‌র = আল্লাহর
            default: out.append(u)
            }
        }
        return String(out)
    }

    /// Substring test on UTF-16 units, so it can match inside a Bangla conjunct or a mark cluster.
    static func has(_ text: String, _ part: String) -> Bool {
        (text as NSString).range(of: part, options: .literal).location != NSNotFound
    }

    static func hasHarakat(_ s: String) -> Bool { s.unicodeScalars.contains { (0x064B...0x0652).contains($0.value) } }

    private func position(_ i: Int) -> (Int, Int)? {
        guard i >= 0, i < Self.total else { return nil }
        var lo = 1, hi = 114
        while lo < hi { let m = (lo + hi + 1) / 2; if start[m] <= i { lo = m } else { hi = m - 1 } }
        return (lo, i - start[lo] + 1)
    }

    /// Āyāt whose Arabic contains `query` (harakat optional: matches with your harakat come first).
    /// `phonetic`: the query came from Latin letters, so similar letters match too (see `loose`).
    public func searchArabic(_ query: String, phonetic: Bool = false, limit: Int = 300) -> (hits: [Passage], total: Int) {
        let q = Self.bare(query)
        guard q.count >= 2 else { return ([], 0) }
        let qa = Self.bare(query, daggerAlif: true)
        let ql = Self.loose(qa)
        let typed = Self.hasHarakat(query) ? Self.light(query) : nil
        // Order: your harakat too → the same letters → (typed in Latin) similar letters, e.g. rahmaan → ٱلرَّحْمَٰنِ.
        var exact: [Int] = [], letters: [Int] = [], similar: [Int] = []
        for (i, t) in bareArabic.enumerated() {
            if Self.has(t.0, q) || Self.has(t.1, q) || Self.has(t.1, qa) {
                if let typed, Self.has(lightArabic[i], typed) { exact.append(i) } else { letters.append(i) }
            } else if phonetic && ql.count >= 3 && (Self.has(looseArabic[i].0, ql) || Self.has(looseArabic[i].1, ql)) {
                similar.append(i)
            }
        }
        let all = exact + letters + similar
        return (all.prefix(limit).compactMap { position($0).map { Passage(surah: $0.0, from: $0.1, to: $0.1) } }, all.count)
    }

    /// Āyāt whose translation contains every word of `query` (case and accents ignored).
    public func searchTranslation(_ id: String, _ query: String, limit: Int = 300) -> (hits: [Passage], total: Int) {
        let words = Self.foldText(query)
            .split(whereSeparator: { $0.isWhitespace || $0.isPunctuation }).map(String.init)
        guard !words.isEmpty, words.joined().count >= 2 else { return ([], 0) }
        lock.lock()
        var folded = foldedTranslation[id]
        lock.unlock()
        if folded == nil {
            guard let t = lines(id) else { return ([], 0) }
            folded = t.map(Self.foldText)
            lock.lock(); foldedTranslation[id] = folded; lock.unlock()
        }
        var hits: [Int] = []
        for (i, t) in folded!.enumerated() where words.allSatisfy({ Self.has(t, $0) }) { hits.append(i) }
        return (hits.prefix(limit).compactMap { position($0).map { Passage(surah: $0.0, from: $0.1, to: $0.1) } }, hits.count)
    }

    /// For showing in the KFGQPC Hafs font, which draws three Tanzil marks as a dotted circle:
    /// small rounded zero ۟ (U+06DF → U+06E0), small low seen ۣ (52:37, → U+06DC), ishmām ۫ (12:11, → U+06EC).
    public static func forKFGQPC(_ s: String) -> String {
        String(String.UnicodeScalarView(s.unicodeScalars.map {
            switch $0.value {
            case 0x06DF: return "\u{06E0}"
            case 0x06E3: return "\u{06DC}"
            case 0x06EB: return "\u{06EC}"
            default: return $0
            }
        }))
    }

    // MARK: Text to insert

    /// "ٱللَّهُ لَآ إِلَٰهَ … (2:255)", then the translation; each āyah has its reference.
    public func text(_ p: Passage, _ mode: QuranInsert, translation id: String?) -> String {
        let trans = mode == .arabic ? nil : id.flatMap { translation($0) != nil ? $0 : nil }
        var out: [String] = []
        for a in p.ayahs {
            let ref = "(\(p.surah):\(a))"
            let ar = arabic(p.surah, a) ?? ""
            let tr = trans.flatMap { translation($0, p.surah, a) }
            switch mode {
            case .arabic: out.append("\(ar) \(ref)")
            case .translation: out.append("\(tr ?? ar) \(ref)")
            case .both: out.append(tr.map { "\(ar)\n\($0) \(ref)" } ?? "\(ar) \(ref)")
            }
        }
        return out.joined(separator: mode == .both ? "\n\n" : " ")
    }
}

// MARK: The user's choice (shared by the keyboard and the app)

extension SharedSettings {
    public static var quranInsert: QuranInsert {
        get { defaults.string(forKey: "quran.insert").flatMap(QuranInsert.init(rawValue:)) ?? .both }
        set { defaults.set(newValue.rawValue, forKey: "quran.insert"); defaults.synchronize() }
    }

    /// Tanzil id of the translation to insert (default Saheeh International).
    public static var quranTranslation: String {
        get { defaults.string(forKey: "quran.translation") ?? "en.sahih" }
        set { defaults.set(newValue, forKey: "quran.translation"); defaults.synchronize() }
    }
}
