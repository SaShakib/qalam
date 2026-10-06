import Foundation

/// One piece of what the user typed.
enum Tok {
    case letter(String)        // an Arabic letter (ب، ع، ي …)
    case vowel(V)              // a i u
    case tanween(V)            // aN iN uN
    case alifKey               // A
    case hamzaKey              // '
    case hamzaSeat(String)     // \hA \hI … (forced seat)
    case taMarbuta             // t'
    case maqsura               // Y
    case shadda, sukun, dagger, maddah, tatweel
    case sep                   // ` (prints nothing, breaks combinations)
    case joiner                // - (prints nothing, marks a prefix)
    case mark(String)          // a combining mark from a \ code
    case literal(String)       // standalone text from a \ code
    case raw(String)           // a key with no meaning: passed through

    var isVowel: Bool {
        switch self {
        case .vowel, .tanween: return true
        default: return false
        }
    }
}

/// Backslash codes: Qur'anic marks, waqf signs, phrases.
enum Codes {
    enum Code {
        case mark(String)
        case standalone(String)
        case seat(String)
        case phrase(everyday: String, quran: String)
    }

    static let table: [String: Code] = [
        // marks on the letter before
        "0": .mark(AR.silent),            // ۟  silent letter
        "w": .mark(AR.smallWaw),          // ۥ
        "y": .mark(AR.smallYa),           // ۦ
        "mi": .mark(AR.iqlabMeem),        // ۢ  iqlāb
        "ml": .mark(AR.lowMeem),          // ۭ
        // waqf signs
        "mm": .mark("\u{06D8}"),          // lāzim
        "la": .mark("\u{06D9}"),          // lā
        "j": .mark("\u{06DA}"),           // jāʾiz
        "sl": .mark("\u{06D6}"),          // ṣilā
        "ql": .mark("\u{06D7}"),          // qilā
        "mu": .mark("\u{06DB}"),          // muʿānaqa
        "sk": .mark("\u{06DC}"),          // sakta
        // symbols
        "sj": .standalone("\u{06E9}"),    // ۩ sajda
        "hz": .standalone("\u{06DE}"),    // ۞ rubʿ / ḥizb
        "saw": .standalone("\u{FDFA}"),   // ﷺ
        // forced hamza seats
        "hA": .seat(AR.hamzaOnAlif),
        "hI": .seat(AR.hamzaUnderAlif),
        "hW": .seat(AR.hamzaOnWaw),
        "hY": .seat(AR.hamzaOnYa),
        "h0": .seat(AR.hamza),
        // phrases
        "bism": .phrase(everyday: "بِسْمِ اللَّهِ الرَّحْمَٰنِ الرَّحِيمِ",
                        quran: "بِسْمِ ٱللَّهِ ٱلرَّحْمَٰنِ ٱلرَّحِيمِ"),
        "swt": .phrase(everyday: "سُبْحَانَهُ وَتَعَالَى", quran: "سُبْحَانَهُ وَتَعَالَىٰ"),
        "as": .phrase(everyday: "عَلَيْهِ السَّلَامُ", quran: "عَلَيْهِ ٱلسَّلَامُ"),
        "ra": .phrase(everyday: "رَضِيَ اللَّهُ عَنْهُ", quran: "رَضِيَ ٱللَّهُ عَنْهُ"),
    ]

    static let maxLen = 4
}

enum Tokenizer {
    static let digraphs: [String: String] = [
        "th": "ث", "dh": "ذ", "kh": "خ", "sh": "ش", "gh": "غ", "zh": "ز",
    ]

    static let singles: [Character: String] = [
        "b": "ب", "t": "ت", "j": "ج", "H": "ح", "d": "د", "r": "ر", "z": "ز", "s": "س",
        "S": "ص", "D": "ض", "T": "ط", "Z": "ظ", "e": "ع", "E": "ع", "g": "غ", "f": "ف",
        "q": "ق", "k": "ك", "l": "ل", "m": "م", "n": "ن", "N": "ن", "h": "ه", "w": "و", "y": "ي",
        // forgiving capitals (no special meaning)
        "B": "ب", "J": "ج", "R": "ر", "G": "غ", "F": "ف", "Q": "ق", "K": "ك", "L": "ل", "M": "م", "W": "و",
    ]

    static func vowel(_ c: Character) -> V? {
        switch c {
        case "a": return .a
        case "i", "I": return .i
        case "u", "U": return .u
        default: return nil
        }
    }

    static func tokenize(_ input: String, _ o: Options) -> [Tok] {
        let c = Array(input)
        var out: [Tok] = []
        var i = 0
        while i < c.count {
            let ch = c[i]

            if ch == "\\" {
                var j = i + 1
                while j < c.count, c[j].isASCII, c[j].isLetter || c[j].isNumber { j += 1 }
                let run = String(c[(i + 1)..<j])
                if let f = run.first, f.isNumber, f != "0" {
                    let digits = String(run.prefix { $0.isNumber })
                    out.append(.literal("\u{FD3F}" + AR.arabicDigits(digits) + "\u{FD3E}"))
                    i += 1 + digits.count
                    continue
                }
                var found: (Int, Codes.Code)?
                var len = min(run.count, Codes.maxLen)
                while len > 0 {
                    if let code = Codes.table[String(run.prefix(len))] { found = (len, code); break }
                    len -= 1
                }
                if let f = found {
                    switch f.1 {
                    case .mark(let m): out.append(.mark(m))
                    case .standalone(let s): out.append(.literal(s))
                    case .seat(let s): out.append(.hamzaSeat(s))
                    case .phrase(let e, let q): out.append(.literal(o.style == .quran ? q : e))
                    }
                    i += 1 + f.0
                } else {
                    out.append(.raw("\\"))
                    i += 1
                }
                continue
            }

            if i + 1 < c.count {
                let pair = String(ch) + String(c[i + 1])
                if pair == "t'" { out.append(.taMarbuta); i += 2; continue }
                if let l = digraphs[pair] { out.append(.letter(l)); i += 2; continue }
            }

            if let v = vowel(ch) {
                if i + 1 < c.count, c[i + 1] == "N" {
                    out.append(.tanween(v)); i += 2
                } else {
                    out.append(.vowel(v)); i += 1
                }
                continue
            }

            switch ch {
            case "A": out.append(.alifKey)
            case "'": out.append(.hamzaKey)
            case "Y": out.append(.maqsura)
            case "~": out.append(.shadda)
            case "o", "O": out.append(.sukun)
            case "^": out.append(.dagger)
            case "=": out.append(.maddah)
            case "_": out.append(.tatweel)
            case "`": out.append(.sep)
            case "-": out.append(.joiner)
            default:
                if let l = singles[ch] { out.append(.letter(l)) } else { out.append(.raw(String(ch))) }
            }
            i += 1
        }
        return out
    }
}
