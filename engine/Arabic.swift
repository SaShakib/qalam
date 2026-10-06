import Foundation

/// Arabic letters and marks used by the engine.
enum AR {
    // Harakat
    static let fatha = "\u{064E}"
    static let damma = "\u{064F}"
    static let kasra = "\u{0650}"
    static let fathatan = "\u{064B}"
    static let dammatan = "\u{064C}"
    static let kasratan = "\u{064D}"
    static let shadda = "\u{0651}"
    static let sukun = "\u{0652}"
    static let quranSukun = "\u{06E1}"   // Madinah-mushaf style sukūn
    static let dagger = "\u{0670}"       // superscript (dagger) alif
    static let maddah = "\u{0653}"

    // Letters
    static let alif = "\u{0627}"
    static let alifWasla = "\u{0671}"
    static let alifMadda = "\u{0622}"
    static let hamza = "\u{0621}"
    static let hamzaOnAlif = "\u{0623}"
    static let hamzaUnderAlif = "\u{0625}"
    static let hamzaOnWaw = "\u{0624}"
    static let hamzaOnYa = "\u{0626}"
    static let taMarbuta = "\u{0629}"
    static let alifMaqsura = "\u{0649}"
    static let tatweel = "\u{0640}"
    static let lam = "\u{0644}"
    static let waw = "\u{0648}"
    static let ya = "\u{064A}"

    // Qur'anic marks
    static let silent = "\u{06DF}"
    static let smallWaw = "\u{06E5}"
    static let smallYa = "\u{06E6}"
    static let iqlabMeem = "\u{06E2}"
    static let lowMeem = "\u{06ED}"

    /// Marks that already say how the letter is read, so it never gets an automatic sukūn.
    static let sukunBlockers: Set<String> = [dagger, maddah, silent, smallWaw, smallYa, iqlabMeem, lowMeem]

    static let sunLetters: Set<String> = ["ت", "ث", "د", "ذ", "ر", "ز", "س", "ش", "ص", "ض", "ط", "ظ", "ل", "ن"]

    /// Letters that never join to the letter after them.
    static let nonJoining: Set<String> = ["ا", "أ", "إ", "آ", "ٱ", "د", "ذ", "ر", "ز", "و", "ؤ", "ء", "ة", "ى"]

    /// Scalars removed in "no harakat" mode.
    static let strippable: Set<Unicode.Scalar> = {
        var s = Set<Unicode.Scalar>()
        for v in 0x064B...0x0652 { s.insert(Unicode.Scalar(v)!) }
        s.insert(Unicode.Scalar(0x0670)!)
        s.insert(Unicode.Scalar(0x0653)!)
        s.insert(Unicode.Scalar(0x06E1)!)
        return s
    }()

    static func stripHarakat(_ s: String) -> String {
        var out = String.UnicodeScalarView()
        for u in s.unicodeScalars where !strippable.contains(u) { out.append(u) }
        return String(out)
    }

    static func arabicDigits(_ s: String) -> String {
        String(s.map { c -> Character in
            guard let d = c.wholeNumberValue, c.isASCII else { return c }
            return Character(Unicode.Scalar(0x0660 + d)!)
        })
    }
}

/// A short vowel.
enum V {
    case a, i, u

    var mark: String {
        switch self {
        case .a: return AR.fatha
        case .i: return AR.kasra
        case .u: return AR.damma
        }
    }

    var tanweenMark: String {
        switch self {
        case .a: return AR.fathatan
        case .i: return AR.kasratan
        case .u: return AR.dammatan
        }
    }

    /// The letter that lengthens this vowel.
    var maddLetter: String {
        switch self {
        case .a: return AR.alif
        case .i: return AR.ya
        case .u: return AR.waw
        }
    }
}
