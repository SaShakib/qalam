import Foundation

/// One version of the word offered in the options panel.
public struct Candidate: Equatable, Sendable {
    public enum Kind: String, Sendable {
        case typed      // exactly what was typed
        case seat       // a different alif / hamza
        case plain      // no harakat
        case dagger     // small alif (لِلّٰهِ)
        case sukun      // with (or with less) sukūn
    }
    public let text: String
    public let kind: Kind
    public let label: String

    public init(text: String, kind: Kind, label: String) {
        self.text = text
        self.kind = kind
        self.label = label
    }
}

extension Qalam {
    /// Up to 4 versions of a word for the options panel. Row 0 is always what was typed.
    public static func candidates(_ typed: String, _ o: Options, limit: Int = 4) -> [Candidate] {
        let main = word(typed, o)
        var out: [Candidate] = [Candidate(text: main, kind: .typed, label: "as typed")]
        func add(_ text: String, _ kind: Candidate.Kind, _ label: String) {
            guard !text.isEmpty, !out.contains(where: { $0.text == text }) else { return }
            out.append(Candidate(text: text, kind: kind, label: label))
        }

        let skeleton = isSkeleton(typed, o)
        let seats = seatVariants(main, typed: typed, skeleton: skeleton)
        if let first = seats.first { add(first.0, .seat, first.1) }
        // qaraA → قَرَأ, then قَرَأَ (hamza + fatḥa)
        if typed.hasSuffix("A") && !skeleton { add(word(typed + "a", o), .seat, "أَ hamza + fatḥa") }
        for v in seats.dropFirst() { add(v.0, .seat, v.1) }
        if o.harakat != .none { add(AR.stripHarakat(main), .plain, "no harakat") }
        if let d = daggerVariant(main) { add(d, .dagger, "small alif ـٰ") }
        if o.harakat == .full && !skeleton {
            var alt = o
            switch effectiveSukun(o) {
            case .full: alt.sukun = .smart
            case .smart, .off: alt.sukun = .full
            }
            if o.style == .quran && alt.sukun == .smart { alt.sukun = .off }
            add(word(typed, alt), .sukun, alt.sukun == .full ? "with sukūn" : "less sukūn")
        }
        return Array(out.prefix(limit))
    }

    static func isMarkScalar(_ v: UInt32) -> Bool {
        (0x064B...0x065F).contains(v) || v == 0x0670 || (0x06D6...0x06ED).contains(v) || v == 0x034F || v == 0x200D
    }

    /// A word typed with no vowels at all (it comes out as bare consonants).
    static func isSkeleton(_ typed: String, _ o: Options) -> Bool {
        if o.spellingWords, SpellingWords.lookup(typed, o) != nil { return false }
        return !hasTypedMarks(Tokenizer.tokenize(typed, o))
    }

    /// Other alif / hamza seats: a bare ا at the end → أ; at the start → أ / إ.
    /// In a word typed without vowels the hamza seat can't be worked out, so offer the other seats.
    static func seatVariants(_ s: String, typed: String, skeleton: Bool) -> [(String, String)] {
        let scalars = Array(s.unicodeScalars)
        let bases = scalars.indices.filter { !isMarkScalar(scalars[$0].value) }
        guard !bases.isEmpty else { return [] }
        var out: [(String, String)] = []

        func replacing(_ i: Int, with v: UInt32) -> String {
            var copy = scalars
            copy[i] = Unicode.Scalar(v)!
            return String(String.UnicodeScalarView(copy))
        }
        let names: [UInt32: String] = [0x0623: "أ alif + hamza", 0x0625: "إ hamza below", 0x0621: "ء hamza alone",
                                       0x0626: "ئ hamza on yāʾ", 0x0624: "ؤ hamza on wāw", 0x0622: "آ alif madda"]

        // A plain alif at the end (qaraA → قَرَأ) or at the start (Aslm → أسلم), not the alif of "al-".
        let startsWithArticle = typed.hasPrefix("al") || typed.hasPrefix("Al")
        if let last = bases.last, scalars[last].value == 0x0627, bases.count > 1 {
            out.append((replacing(last, with: 0x0623), names[0x0623]!))
        }
        if let first = bases.first, scalars[first].value == 0x0627, !startsWithArticle {
            // the vowel on the alif decides: kasra → إ, fatḥa / ḍamma → أ, none → both
            let mark = first + 1 < scalars.count ? scalars[first + 1].value : 0
            let seats: [UInt32] = mark == 0x0650 ? [0x0625] : (mark == 0x064E || mark == 0x064F) ? [0x0623] : [0x0623, 0x0625]
            for v in seats { out.append((replacing(first, with: v), names[v]!)) }
        }

        // The first hamza letter of a vowel-less word: offer the other seats.
        let hamzas: [UInt32] = [0x0623, 0x0621, 0x0626, 0x0624, 0x0625]
        if skeleton, let h = bases.first(where: { hamzas.contains(scalars[$0].value) }) {
            let current = scalars[h].value
            for v in hamzas where v != current { out.append((replacing(h, with: v), names[v]!)) }
        }
        return out
    }

    /// لِلَّهِ → لِلّٰهِ, اللَّهُ → اللّٰهُ (fatḥa + shadda on the second lām of لله becomes shadda + small alif).
    static func daggerVariant(_ s: String) -> String? {
        let sc = Array(s.unicodeScalars)
        var out = String.UnicodeScalarView()
        var changed = false
        var i = 0
        while i < sc.count {
            out.append(sc[i])
            if sc[i].value == 0x0644 {
                var j = i + 1
                var marks: [UInt32] = []
                while j < sc.count, [0x064E, 0x0651].contains(sc[j].value) { marks.append(sc[j].value); j += 1 }
                if Set(marks) == [0x064E, 0x0651], j < sc.count, sc[j].value == 0x0647 {
                    out.append(Unicode.Scalar(0x0651)!)
                    out.append(Unicode.Scalar(0x0670)!)
                    changed = true
                    i = j
                    continue
                }
            }
            i += 1
        }
        return changed ? String(out) : nil
    }
}
