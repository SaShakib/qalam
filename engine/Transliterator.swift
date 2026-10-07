import Foundation

/// One written letter with its marks.
struct Unit {
    enum Kind { case consonant, hamza, alif, madda, wasl, taMarbuta, maqsura, articleLam, tatweel, literal, raw }

    var kind: Kind
    var base: String
    var vowel: V?
    var tanween: V?
    var shadda = false
    var sukun = false
    var marks: [String] = []
    var forcedSeat: String?
    var chunkStart = false      // first letter of the word, or first after a "-" prefix
    var wordStart = false       // first letter of the whole word
    var isMadd = false          // و / ي that only lengthen the vowel before them
    var assimilated = false     // article lām before a sun letter

    init(_ kind: Kind, _ base: String) {
        self.kind = kind
        self.base = base
    }

    var hasDagger: Bool { marks.contains(AR.dagger) }
    var blocksSukun: Bool { marks.contains { AR.sukunBlockers.contains($0) } }
    var isBare: Bool { vowel == nil && tanween == nil && !shadda && !sukun && marks.isEmpty }
    var takesVowel: Bool {
        (kind == .consonant || kind == .hamza || kind == .taMarbuta) && vowel == nil && tanween == nil && !sukun
    }
}

public enum Qalam {

    // MARK: Public API

    /// Transliterate one word (what sits between spaces).
    public static func word(_ input: String, _ o: Options) -> String {
        if input.isEmpty { return "" }
        if o.spellingWords, !input.contains("\\"), let s = SpellingWords.lookup(input, o) {
            return finish(s, o)
        }
        let toks = Tokenizer.tokenize(input, o)
        var units = buildUnits(toks)
        applyArticle(&units)
        markMadd(&units)
        seatHamzas(&units)
        insertTanweenAlif(&units)
        if o.harakat == .full { autoSukun(&units) }
        return finish(render(units, o), o)
    }

    /// Transliterate running text: words, spaces, punctuation, numbers.
    public static func text(_ s: String, _ o: Options) -> String {
        var comp = Composer()
        var out = ""
        for c in s {
            let key: Composer.Key = c == " " ? .space : (c.isNewline ? .other : .char(c))
            let (consumed, actions) = comp.handle(key, o)
            for a in actions { if case .commit(let t) = a { out += t } }
            if !consumed { out.append(c) }
        }
        for a in comp.flushActions(o) { if case .commit(let t) = a { out += t } }
        return out
    }

    /// Does this key continue (or start) a word being composed?
    public static func accepts(_ c: Character, buffer: String) -> Bool {
        if c.isASCII && c.isLetter { return true }
        switch c {
        case "'", "\\", "_": return true
        case "`", "^", "=", "~", "-": return !buffer.isEmpty
        default: break
        }
        if c.isASCII, c.isNumber, let r = buffer.range(of: "\\", options: .backwards) {
            let tail = buffer[r.upperBound...]
            return tail.allSatisfy { $0.isNumber } || tail == "h"
        }
        return false
    }

    /// Arabic replacement for a key that is not part of a word (nil = leave it as it is).
    public static func punctuation(_ c: Character, _ o: Options) -> String? {
        switch c {
        case ",": return "،"
        case ";": return "؛"
        case "?": return "؟"
        default: break
        }
        if o.arabicDigits, c.isASCII, c.isNumber { return AR.arabicDigits(String(c)) }
        return nil
    }

    // MARK: Building letters

    static func buildUnits(_ toks: [Tok]) -> [Unit] {
        var chunks: [[Tok]] = [[]]
        for t in toks {
            if case .joiner = t { chunks.append([]) } else { chunks[chunks.count - 1].append(t) }
        }

        var units: [Unit] = []
        var ci = 0
        while ci < chunks.count {
            let ch = chunks[ci]
            let isLast = ci == chunks.count - 1
            let start = units.count

            if !isLast {
                // al-
                if ch.count == 2, case .vowel(.a) = ch[0], case .letter(let l) = ch[1], l == AR.lam {
                    addArticle(&units, start)
                    ci += 1; continue
                }
                // ash-shams (article written as heard)
                if ch.count == 2, case .vowel(.a) = ch[0], case .letter(let x) = ch[1], AR.sunLetters.contains(x),
                   let f = chunks[ci + 1].first, case .letter(let y) = f, y == x {
                    addArticle(&units, start)
                    ci += 1; continue
                }
                // wal-, fal-, bil-, kal-, lil-
                if ch.count == 3, case .letter(let p) = ch[0], ["و", "ف", "ب", "ك", "ل"].contains(p),
                   case .vowel(let v) = ch[1], case .letter(let l) = ch[2], l == AR.lam {
                    var pu = Unit(.consonant, p)
                    pu.vowel = v
                    add(pu, &units, start)
                    if p != AR.lam { add(Unit(.wasl, AR.alif), &units, start) }
                    add(Unit(.articleLam, AR.lam), &units, start)
                    ci += 1; continue
                }
            }

            var rest = ch
            // al + letter inside the word: the article
            if ch.count >= 3, case .vowel(.a) = ch[0], case .letter(let l) = ch[1], l == AR.lam, articleFollows(ch, 2) {
                addArticle(&units, start)
                rest = Array(ch[2...])
            }
            process(rest, &units, start)
            ci += 1
        }
        return units
    }

    static func articleFollows(_ ch: [Tok], _ i: Int) -> Bool {
        guard i < ch.count else { return false }
        switch ch[i] {
        case .letter, .hamzaKey, .hamzaSeat: return true
        case .alifKey: return i + 1 < ch.count && ch[i + 1].isVowel
        default: return false
        }
    }

    static func addArticle(_ units: inout [Unit], _ start: Int) {
        add(Unit(.wasl, AR.alif), &units, start)
        add(Unit(.articleLam, AR.lam), &units, start)
    }

    static func add(_ u: Unit, _ units: inout [Unit], _ chunkStart: Int) {
        var u = u
        u.chunkStart = units.count == chunkStart
        u.wordStart = units.isEmpty
        units.append(u)
    }

    static func process(_ toks: [Tok], _ units: inout [Unit], _ start: Int) {
        var lastWasSep = false
        var k = 0
        while k < toks.count {
            let t = toks[k]
            let next: Tok? = k + 1 < toks.count ? toks[k + 1] : nil
            let inChunk = units.count > start
            let li: Int? = units.isEmpty ? nil : units.count - 1

            switch t {
            case .letter(let L):
                if !lastWasSep, inChunk, let li, units[li].kind == .consonant, units[li].base == L, units[li].isBare {
                    units[li].shadda = true          // doubled letter → shadda
                } else {
                    add(Unit(.consonant, L), &units, start)
                }

            case .vowel(let v):
                if !inChunk {
                    var u = Unit(.wasl, AR.alif)     // word starts with a vowel → hamzat al-waṣl
                    u.vowel = v
                    add(u, &units, start)
                } else if let li, units[li].takesVowel {
                    units[li].vowel = v
                } else {
                    add(Unit(v == .a ? .alif : .consonant, v.maddLetter), &units, start)   // aa ii uu
                }

            case .tanween(let v):
                if inChunk, let li, units[li].takesVowel {
                    units[li].tanween = v
                }

            case .alifKey:
                if let n = next, n.isVowel {
                    add(Unit(.hamza, AR.hamza), &units, start)          // A + haraka = hamza
                } else if case .alifKey? = next, !(k + 2 < toks.count && toks[k + 2].isVowel) {
                    add(Unit(.madda, AR.alifMadda), &units, start)      // AA = آ
                    k += 1
                } else {
                    autoFatha(&units, start)
                    add(Unit(.alif, AR.alif), &units, start)
                }

            case .hamzaKey:
                add(Unit(.hamza, AR.hamza), &units, start)

            case .hamzaSeat(let s):
                var u = Unit(.hamza, AR.hamza)
                u.forcedSeat = s
                add(u, &units, start)

            case .taMarbuta:
                add(Unit(.taMarbuta, AR.taMarbuta), &units, start)

            case .maqsura:
                autoFatha(&units, start)
                add(Unit(.maqsura, AR.alifMaqsura), &units, start)

            case .shadda:
                if let li { units[li].shadda = true }
            case .sukun:
                if let li { units[li].sukun = true }
            case .dagger:
                if let li { units[li].marks.append(AR.dagger) }
            case .maddah:
                if let li { units[li].marks.append(AR.maddah) }
            case .mark(let m):
                if let li { units[li].marks.append(m) } else { add(Unit(.literal, m), &units, start) }

            case .tatweel:
                add(Unit(.tatweel, AR.tatweel), &units, start)
            case .sep:
                lastWasSep = true
                k += 1
                continue
            case .joiner:
                break
            case .literal(let s):
                add(Unit(.literal, s), &units, start)
            case .raw(let s):
                add(Unit(.raw, s), &units, start)
            }
            lastWasSep = false
            k += 1
        }
    }

    /// `kAtaba`: a letter right before an alif gets its fatḥa automatically.
    static func autoFatha(_ units: inout [Unit], _ start: Int) {
        guard units.count > start else { return }
        let li = units.count - 1
        let u = units[li]
        // a madd letter (كَتَبُو + ا) already carries its vowel
        if li > 0, (u.base == AR.waw && units[li - 1].vowel == .u) || (u.base == AR.ya && units[li - 1].vowel == .i) { return }
        if (u.kind == .consonant || u.kind == .hamza), u.vowel == nil, u.tanween == nil, !u.sukun {
            units[li].vowel = .a
        }
    }

    // MARK: Rules

    static func applyArticle(_ units: inout [Unit]) {
        for i in units.indices where units[i].kind == .articleLam {
            if i + 1 < units.count, units[i + 1].kind == .consonant, AR.sunLetters.contains(units[i + 1].base) {
                units[i].assimilated = true
                units[i + 1].shadda = true
            }
        }
    }

    static func markMadd(_ units: inout [Unit]) {
        guard units.count > 1 else { return }
        for i in 1..<units.count {
            let u = units[i], p = units[i - 1]
            guard u.kind == .consonant, u.vowel == nil, u.tanween == nil, !u.shadda, !u.sukun else { continue }
            if (u.base == AR.waw && p.vowel == .u) || (u.base == AR.ya && p.vowel == .i) {
                units[i].isMadd = true
            }
        }
    }

    enum PrevV { case a, i, u, sukun, longA, longI, longU }

    static func prevVowel(_ p: Unit) -> PrevV {
        if p.kind == .alif || p.kind == .madda { return .longA }
        if p.isMadd { return p.base == AR.waw ? .longU : .longI }
        if p.hasDagger { return .longA }
        switch p.vowel ?? p.tanween {
        case .a?: return .a
        case .i?: return .i
        case .u?: return .u
        case nil: return .sukun
        }
    }

    static func seatHamzas(_ units: inout [Unit]) {
        var i = 0
        while i < units.count {
            if units[i].kind == .hamza { seat(i, &units) }
            i += 1
        }
    }

    static func seat(_ i: Int, _ units: inout [Unit]) {
        let u = units[i]
        if let f = u.forcedSeat { units[i].base = f; return }

        let own = u.vowel ?? u.tanween
        let prev: Unit? = i > 0 ? units[i - 1] : nil
        let initial = u.chunkStart || prev == nil || prev?.kind == .articleLam || prev?.kind == .wasl
        let final = i == units.count - 1
        var seat: String

        if initial {
            seat = own == .i ? AR.hamzaUnderAlif : AR.hamzaOnAlif
        } else if let p = prev {
            let pv = prevVowel(p)
            if final {
                switch pv {
                case .i: seat = AR.hamzaOnYa
                case .u: seat = AR.hamzaOnWaw
                case .a: seat = AR.hamzaOnAlif
                default: seat = AR.hamza
                }
                // شَيْئًا but جُزْءًا: before the alif of tanwīn the hamza sits on a tooth if the letter before joins
                if seat == AR.hamza, u.tanween == .a, pv == .sukun || pv == .longI, !AR.nonJoining.contains(p.base) {
                    seat = AR.hamzaOnYa
                }
            } else {
                if own == .i || pv == .i || pv == .longI || (p.base == AR.ya && pv == .sukun) {
                    seat = AR.hamzaOnYa
                } else if pv == .longA && own == .a {
                    seat = AR.hamza
                } else if pv == .longU {
                    seat = AR.hamza
                } else if own == .u || pv == .u {
                    seat = AR.hamzaOnWaw
                } else {
                    seat = AR.hamzaOnAlif
                }
            }
        } else {
            seat = AR.hamzaOnAlif
        }
        units[i].base = seat

        // أَ + ا = آ
        if seat == AR.hamzaOnAlif, u.vowel == .a, i + 1 < units.count, units[i + 1].kind == .alif {
            units[i].kind = .madda
            units[i].base = AR.alifMadda
            units[i].vowel = nil
            units.remove(at: i + 1)
        }
    }

    static func insertTanweenAlif(_ units: inout [Unit]) {
        var i = 0
        while i < units.count {
            let u = units[i]
            if u.tanween == .a {
                let next: Unit? = i + 1 < units.count ? units[i + 1] : nil
                let afterLongA = i > 0 && (units[i - 1].kind == .alif || units[i - 1].kind == .madda)
                let skip = u.kind == .taMarbuta
                    || next?.kind == .maqsura
                    || next?.kind == .alif
                    || (u.kind == .hamza && u.base == AR.hamza && afterLongA)
                if !skip { units.insert(Unit(.alif, AR.alif), at: i + 1) }
            }
            i += 1
        }
    }

    static func autoSukun(_ units: inout [Unit]) {
        for i in units.indices {
            let u = units[i]
            guard u.kind == .consonant || u.kind == .hamza || u.kind == .articleLam else { continue }
            if u.vowel != nil || u.tanween != nil || u.shadda || u.sukun || u.isMadd || u.assimilated || u.blocksSukun { continue }
            units[i].sukun = true
        }
    }

    // MARK: Output

    /// Inserts U+034F (combining grapheme joiner) between the two lāms of ل‌ل‌ه (ignoring marks),
    /// so fonts don't replace the word with their built-in "Allah" ligature.
    public static func blockAllahLigature(_ s: String) -> String {
        let scalars = Array(s.unicodeScalars)
        func isMark(_ u: Unicode.Scalar) -> Bool {
            let v = u.value
            return (0x064B...0x065F).contains(v) || v == 0x0670 || (0x06D6...0x06ED).contains(v) || v == 0x034F
        }
        // positions of base letters
        let bases = scalars.indices.filter { !isMark(scalars[$0]) }
        var insertBefore = Set<Int>()
        for k in 0..<max(0, bases.count - 2) {
            let a = scalars[bases[k]].value, b = scalars[bases[k + 1]].value, c = scalars[bases[k + 2]].value
            if a == 0x0644 && b == 0x0644 && c == 0x0647 { insertBefore.insert(bases[k + 1]) }
        }
        if insertBefore.isEmpty { return s }
        var out = String.UnicodeScalarView()
        for (i, u) in scalars.enumerated() {
            if insertBefore.contains(i) { out.append(Unicode.Scalar(0x034F)!) }
            out.append(u)
        }
        return String(out)
    }

    static func render(_ units: [Unit], _ o: Options) -> String {
        var s = ""
        for u in units {
            if u.kind == .wasl {
                if o.style == .quran {
                    s += AR.alifWasla
                } else {
                    s += AR.alif
                    if u.wordStart, let v = u.vowel { s += v.mark }
                }
                s += u.marks.joined()
                continue
            }
            s += u.base
            if u.shadda { s += AR.shadda }
            if let v = u.vowel {
                s += v.mark
            } else if let t = u.tanween {
                s += t.tanweenMark
            } else if u.sukun {
                s += AR.sukun
            }
            s += u.marks.joined()
        }
        return s
    }

    static func finish(_ s: String, _ o: Options) -> String {
        let s = o.blockAllahLigature ? blockAllahLigature(s) : s
        if o.harakat == .none { return AR.stripHarakat(s) }
        if o.style == .quran && o.quranSmallSukun {
            let from = AR.sukun.unicodeScalars.first!, to = AR.quranSukun.unicodeScalars.first!
            var out = String.UnicodeScalarView()
            for u in s.unicodeScalars { out.append(u == from ? to : u) }
            return String(out)
        }
        return s
    }
}
