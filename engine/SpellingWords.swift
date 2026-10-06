import Foundation

/// Words whose spelling is not what they sound like (hidden alif, اللَّه …).
/// Typed as heard, they come out spelled as written.
enum SpellingWords {
    struct Entry {
        let everyday: String
        let quran: String
        let plain: String
        /// Open stems take a case ending typed after them (allaAh + u → اللَّهُ).
        let open: Bool
    }

    /// Keys are in normalized form (see `normalize`).
    static let table: [String: Entry] = {
        var t: [String: Entry] = [:]
        func put(_ keys: [String], _ e: String, _ q: String, _ p: String, open: Bool) {
            for k in keys { t[k] = Entry(everyday: e, quran: q, plain: p, open: open) }
        }
        put(["allaah", "allah"], "اللَّه", "ٱللَّه", "الله", open: true)
        put(["wallaah", "wallah"], "وَاللَّه", "وَٱللَّه", "والله", open: true)
        put(["billaah", "billah"], "بِاللَّه", "بِٱللَّه", "بالله", open: true)
        put(["lillaah", "lillah"], "لِلَّه", "لِلَّه", "لله", open: true)
        put(["tallaah"], "تَاللَّه", "تَٱللَّه", "تالله", open: true)
        put(["allaahumma", "allahumma"], "اللَّهُمَّ", "ٱللَّهُمَّ", "اللهم", open: false)
        // relative pronouns are written with one lām
        put(["alladhii"], "الَّذِي", "ٱلَّذِي", "الذي", open: false)
        put(["allatii"], "الَّتِي", "ٱلَّتِي", "التي", open: false)
        put(["alladhiina", "alladhiin"], "الَّذِينَ", "ٱلَّذِينَ", "الذين", open: false)
        put(["'ilaah"], "إِلَٰه", "إِلَٰه", "إله", open: true)
        put(["haadhaa"], "هَٰذَا", "هَٰذَا", "هذا", open: false)
        put(["haadhihi"], "هَٰذِهِ", "هَٰذِهِ", "هذه", open: false)
        put(["dhaalik"], "ذَٰلِك", "ذَٰلِك", "ذلك", open: true)
        put(["dhaalikum"], "ذَٰلِكُمْ", "ذَٰلِكُمْ", "ذلكم", open: false)
        put(["kadhaalik"], "كَذَٰلِك", "كَذَٰلِك", "كذلك", open: true)
        put(["haakadhaa"], "هَٰكَذَا", "هَٰكَذَا", "هكذا", open: false)
        put(["laakin"], "لَٰكِن", "لَٰكِن", "لكن", open: true)
        put(["laakinna"], "لَٰكِنَّ", "لَٰكِنَّ", "لكن", open: false)
        put(["haa'ulaa'"], "هَٰؤُلَاء", "هَٰٓؤُلَآء", "هؤلاء", open: true)
        put(["'ulaa'ik", "'uulaa'ik"], "أُولَٰئِك", "أُو۟لَٰٓئِك", "أولئك", open: true)
        put(["alraHmaan", "al-raHmaan", "ar-raHmaan"], "الرَّحْمَٰن", "ٱلرَّحْمَٰن", "الرحمن", open: true)
        put(["raHmaan"], "رَحْمَٰن", "رَحْمَٰن", "رحمن", open: true)
        return t
    }()

    static let suffixes: [(String, V, Bool)] = [
        ("aN", .a, true), ("iN", .i, true), ("uN", .u, true),
        ("a", .a, false), ("i", .i, false), ("u", .u, false),
    ]

    static func lookup(_ input: String, _ o: Options) -> String? {
        let n = normalize(input)
        if let e = table[n] { return build(e, nil, false, o) }
        for (suf, v, tan) in suffixes where n.hasSuffix(suf) {
            if let e = table[String(n.dropLast(suf.count))], e.open { return build(e, v, tan, o) }
        }
        return nil
    }

    static func build(_ e: Entry, _ v: V?, _ tanween: Bool, _ o: Options) -> String {
        if o.harakat == .none { return e.plain + (tanween && v == .a ? AR.alif : "") }
        var s = o.style == .quran ? e.quran : e.everyday
        if let v {
            s += tanween ? v.tanweenMark : v.mark
            if tanween && v == .a { s += AR.alif }
        } else if e.open && o.harakat == .full {
            s += AR.sukun
        }
        return s
    }

    /// Different ways of typing the same word → one key.
    /// `aA`→`aa`, `iy`→`ii`, `uw`→`uu`, `A`+haraka→`'`.
    static func normalize(_ s: String) -> String {
        let c = Array(s)
        var out = ""
        for (i, ch) in c.enumerated() {
            let prev: Character? = i > 0 ? c[i - 1] : nil
            let next: Character? = i + 1 < c.count ? c[i + 1] : nil
            let nextIsVowel = next.map { "aiu".contains($0) } ?? false
            switch ch {
            case "A":
                if nextIsVowel { out.append("'") } else if prev == "a" { out.append("a") } else { out.append("A") }
            case "y" where prev == "i" && !(next.map { "aiuy".contains($0) } ?? false):
                out.append("i")
            case "w" where prev == "u" && !(next.map { "aiuw".contains($0) } ?? false):
                out.append("u")
            case "E":
                out.append("e")
            default:
                out.append(ch)
            }
        }
        return out
    }
}
