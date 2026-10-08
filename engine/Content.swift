import Foundation

// Teaching content shown in the app: the letter table, key tables, word lists and drills.
// Arabic is never stored here: it is always produced from the typed form by the engine,
// so what the app shows is exactly what the keyboard will type.

public struct LetterRow: Identifiable, Sendable {
    public var id: String { arabic }
    public let arabic: String
    public let name: String
    public let keys: [String]
    public let example: String
    public let meaning: String
    public let drills: [WordItem]
}

public struct KeyRow: Identifiable, Sendable {
    public var id: String { name }
    public let shows: String
    public let name: String
    public let keys: String
    public let example: String
}

public struct WordItem: Identifiable, Hashable, Sendable {
    public var id: String { latin }
    public let latin: String
    public let meaning: String
}

public struct WordGroup: Identifiable, Sendable {
    public let id: String
    public let title: String
    public let note: String
    public let style: Style
    public let items: [WordItem]
}

public struct PracticeItem: Identifiable, Hashable, Sendable {
    public var id: String { latin + label }
    public let latin: String
    public let label: String
    public let meaning: String
    public let style: Style
}

public enum Content {

    // MARK: Letters

    static func L(_ a: String, _ name: String, _ keys: [String], _ ex: String, _ meaning: String,
                  _ drills: [(String, String)]) -> LetterRow {
        LetterRow(arabic: a, name: name, keys: keys, example: ex, meaning: meaning,
                  drills: drills.map { WordItem(latin: $0.0, meaning: $0.1) })
    }

    public static let letters: [LetterRow] = [
        L("ا", "alif", ["A", "aa"], "kaAtaba", "he wrote to",
          [("baAbuN", "a door"), ("qaAla", "he said"), ("naAruN", "fire")]),
        L("ب", "bāʾ", ["b"], "baAbuN", "a door",
          [("baytuN", "a house"), ("kalbuN", "a dog"), ("bintuN", "a girl")]),
        L("ت", "tāʾ", ["t"], "tamruN", "dates",
          [("tamruN", "dates"), ("bintuN", "a girl"), ("taAjuN", "a crown")]),
        L("ث", "thāʾ", ["th"], "thawbuN", "a garment",
          [("thawbuN", "a garment"), ("thalaAthat'uN", "three"), ("mithluN", "like")]),
        L("ج", "jīm", ["j"], "jabaluN", "a mountain",
          [("jabaluN", "a mountain"), ("rajuluN", "a man"), ("masjiduN", "a mosque")]),
        L("ح", "ḥāʾ", ["H"], "HubbuN", "love",
          [("HubbuN", "love"), ("SabaAHuN", "morning"), ("naHnu", "we")]),
        L("خ", "khāʾ", ["kh"], "khubzuN", "bread",
          [("khubzuN", "bread"), ("AakhuN", "a brother"), ("kharaja", "he went out")]),
        L("د", "dāl", ["d"], "daAruN", "a home",
          [("daAruN", "a home"), ("yaduN", "a hand"), ("waladuN", "a boy")]),
        L("ذ", "dhāl", ["dh"], "dhahabuN", "gold",
          [("dhahabuN", "gold"), ("haAdhaA", "this"), ("Aakhadha", "he took")]),
        L("ر", "rāʾ", ["r"], "rajuluN", "a man",
          [("rajuluN", "a man"), ("qamaruN", "a moon"), ("naAruN", "fire")]),
        L("ز", "zāy", ["z", "zh"], "zaytuN", "oil",
          [("zaytuN", "oil"), ("khubzuN", "bread"), ("zawjuN", "a husband")]),
        L("س", "sīn", ["s"], "salaAmuN", "peace",
          [("salaAmuN", "peace"), ("masjiduN", "a mosque"), ("samiea", "he heard")]),
        L("ش", "shīn", ["sh"], "shamsuN", "a sun",
          [("shamsuN", "a sun"), ("shariba", "he drank"), ("easharat'uN", "ten")]),
        L("ص", "ṣād", ["S"], "SabruN", "patience",
          [("SabruN", "patience"), ("SadiyquN", "a friend"), ("qiSSat'uN", "a story")]),
        L("ض", "ḍād", ["D"], "DayfuN", "a guest",
          [("DayfuN", "a guest"), ("AarDuN", "land"), ("maraDuN", "an illness")]),
        L("ط", "ṭāʾ", ["T"], "TaAlibuN", "a student",
          [("TaAlibuN", "a student"), ("TariyquN", "a road"), ("qiTTuN", "a cat")]),
        L("ظ", "ẓāʾ", ["Z"], "ZulmuN", "injustice",
          [("ZulmuN", "injustice"), ("naZara", "he looked"), ("ZuhruN", "noon")]),
        L("ع", "ʿayn", ["e"], "eilmuN", "knowledge",
          [("eilmuN", "knowledge"), ("maea", "with"), ("sabeat'uN", "seven")]),
        L("غ", "ghayn", ["g", "gh"], "gafara", "he forgave",
          [("gafara", "he forgave"), ("lugat'uN", "a language"), ("gurfat'uN", "a room")]),
        L("ف", "fāʾ", ["f"], "fajruN", "dawn",
          [("fajruN", "dawn"), ("fiy", "in"), ("SayfuN", "summer")]),
        L("ق", "qāf", ["q"], "qalamuN", "a pen",
          [("qalamuN", "a pen"), ("qalbuN", "a heart"), ("SadiyquN", "a friend")]),
        L("ك", "kāf", ["k"], "kalbuN", "a dog",
          [("kalbuN", "a dog"), ("kitaAbuN", "a book"), ("malikuN", "a king")]),
        L("ل", "lām", ["l"], "layluN", "night",
          [("layluN", "night"), ("qalamuN", "a pen"), ("laA", "no")]),
        L("م", "mīm", ["m"], "maAAuN", "water",
          [("maAAuN", "water"), ("yawmuN", "a day"), ("eilmuN", "knowledge")]),
        L("ن", "nūn", ["n"], "nuwruN", "light",
          [("nuwruN", "light"), ("bintuN", "a girl"), ("Aanta", "you")]),
        L("ه", "hāʾ", ["h"], "huwa", "he",
          [("huwa", "he"), ("wajhuN", "a face"), ("qahwat'uN", "coffee")]),
        L("و", "wāw", ["w"], "waladuN", "a boy",
          [("waladuN", "a boy"), ("yawmuN", "a day"), ("nuwruN", "light")]),
        L("ي", "yāʾ", ["y"], "yawmuN", "a day",
          [("yawmuN", "a day"), ("baytuN", "a house"), ("fiy", "in")]),
        L("ء", "hamza", ["'", "A + haraka"], "saAala", "he asked",
          [("saAala", "he asked"), ("qaraAa", "he read"), ("mu'minuN", "a believer"), ("shay'uN", "a thing")]),
        L("ة", "tāʾ marbūṭa", ["t'"], "madrasat'uN", "a school",
          [("madrasat'uN", "a school"), ("thalaAthat'uN", "three"), ("SalaAt'uN", "prayer")]),
        L("ى", "alif maqṣūra", ["Y"], "ealaY", "on",
          [("ealaY", "on"), ("mataY", "when?"), ("hudaNY", "guidance")]),
        L("آ", "alif madda", ["AA"], "AAmana", "he believed",
          [("AAmana", "he believed"), ("qurAAnuN", "the Qur'an"), ("AAkhiruN", "last")]),
    ]

    // MARK: Starter practice (shown on first launch)

    /// Twenty real words, each teaching one thing that is not obvious.
    public static let starter: [PracticeItem] = [
        ("kitaAbuN", "a book", "letter, then its haraka · A = alif (long ā) · uN = tanwīn"),
        ("maktabuN", "a desk", "a letter with no haraka between two others gets sukūn by itself (كْ)"),
        ("mudarrisuN", "a teacher", "type a letter twice → shadda (rr = رّ)"),
        ("yaquwlu", "he says", "uw = long ū (uu works too)"),
        ("dhahaba", "he went", "dh = ذ"),
        ("thalaAthat'uN", "three", "th = ث · t' = ة"),
        ("khubzuN", "bread", "kh = خ · z = ز"),
        ("shamsuN", "a sun", "sh = ش"),
        ("gafara", "he forgave", "g = غ"),
        ("HubbuN", "love", "Shift: H = ح (h = ه) · bb = shadda"),
        ("SabruN", "patience", "Shift: S = ص (s = س)"),
        ("DayfuN", "a guest", "Shift: D = ض (d = د)"),
        ("TaAlibuN", "a student", "Shift: T = ط (t = ت)"),
        ("ZulmuN", "injustice", "Shift: Z = ظ (z = ز)"),
        ("eilmuN", "knowledge", "e = ع (ʿayn), then its haraka"),
        ("saAala", "he asked", "A + haraka = hamza; the seat (أ) is chosen for you"),
        ("mu'minuN", "a believer", "' = hamza with no haraka (ؤْ)"),
        ("ealaY", "on", "Y = ى (alif maqṣūra)"),
        ("AAmana", "he believed", "AA = آ (alif madda)"),
        ("alshamsu", "the sun", "al + sun letter → the lām is silent, the next letter gets shadda"),
    ].map { PracticeItem(latin: $0.0, label: $0.2, meaning: $0.1, style: .everyday) }

    // MARK: Keys

    public static let harakat: [KeyRow] = [
        KeyRow(shows: "ـَ", name: "fatḥa", keys: "a", example: "kataba"),
        KeyRow(shows: "ـِ", name: "kasra", keys: "i", example: "bismi"),
        KeyRow(shows: "ـُ", name: "ḍamma", keys: "u", example: "kutubuN"),
        KeyRow(shows: "ـٌ  ـٍ  ـً", name: "tanwīn", keys: "uN   iN   aN", example: "kitaAbaN"),
        KeyRow(shows: "ـّ", name: "shadda", keys: "type the letter twice", example: "rabbi"),
        KeyRow(shows: "ـْ", name: "sukūn", keys: "automatic at a stop · o by hand", example: "maktab"),
        KeyRow(shows: "ـَا", name: "long ā", keys: "aA  or  aa", example: "qaAla"),
        KeyRow(shows: "ـِي", name: "long ī", keys: "iy  or  ii", example: "kabiyruN"),
        KeyRow(shows: "ـُو", name: "long ū", keys: "uw  or  uu", example: "yaquwlu"),
        KeyRow(shows: "ـَيْ  ـَوْ", name: "ay / aw", keys: "ay   aw", example: "baytuN"),
    ]

    public static let symbols: [KeyRow] = [
        KeyRow(shows: "أ إ ؤ ئ ء", name: "hamza seat", keys: "' or A + haraka (seat is automatic)", example: "suAaAluN"),
        KeyRow(shows: "اِ", name: "hamzat al-waṣl", keys: "start a word with a bare vowel", example: "ismuN"),
        KeyRow(shows: "الْ  الشَّ", name: "the article", keys: "al…  (wal- bil- lil- fal- kal-)", example: "alshamsu"),
        KeyRow(shows: "ـٰ", name: "dagger alif", keys: "^", example: "ha^dhaa"),
        KeyRow(shows: "ـٓ", name: "maddah", keys: "=", example: "jaA='a"),
        KeyRow(shows: "—", name: "separator", keys: "`  (s`h = سه, not ش)", example: "Aas`halu"),
        KeyRow(shows: "—", name: "joiner", keys: "-  (after a prefix)", example: "wa-qaAla"),
        KeyRow(shows: "، ؛ ؟", name: "punctuation", keys: ",  ;  ?", example: "hal fahimta?"),
        KeyRow(shows: "۞ ۚ ﴿٧﴾", name: "Qur'an codes", keys: "\\hz  \\j  \\7  \\bism  \\saw", example: "\\saw"),
        KeyRow(shows: "Esc", name: "undo Arabic", keys: "Esc gives back the English letters", example: ""),
    ]

    // MARK: Words

    static func W(_ latin: String, _ meaning: String) -> WordItem { WordItem(latin: latin, meaning: meaning) }

    public static let groups: [WordGroup] = [
        WordGroup(id: "first", title: "First words", note: "Everyday nouns. Remember: capital N is tanwīn.", style: .everyday, items: [
            W("kitaAbuN", "a book"), W("qalamuN", "a pen"), W("baytuN", "a house"), W("baAbuN", "a door"),
            W("waladuN", "a boy"), W("bintuN", "a girl"), W("rajuluN", "a man"), W("imraAat'uN", "a woman"),
            W("AabuN", "a father"), W("AummuN", "a mother"), W("AakhuN", "a brother"), W("AukhtuN", "a sister"),
            W("SadiyquN", "a friend"), W("TaAlibuN", "a student"), W("mudarrisuN", "a teacher"), W("madrasat'uN", "a school"),
            W("masjiduN", "a mosque"), W("madiynat'uN", "a city"), W("TariyquN", "a road"), W("sayyaArat'uN", "a car"),
            W("maAAuN", "water"), W("khubzuN", "bread"), W("TaeaAmuN", "food"), W("qahwat'uN", "coffee"),
            W("shaAyuN", "tea"), W("shamsuN", "a sun"), W("qamaruN", "a moon"), W("yawmuN", "a day"),
            W("laylat'uN", "a night"), W("waqtuN", "time"), W("lugat'uN", "a language"), W("kalimat'uN", "a word"),
            W("SabaAHuN", "morning"), W("masaaAuN", "evening"), W("qalbuN", "a heart"), W("eaynuN", "an eye"),
            W("yaduN", "a hand"), W("ismuN", "a name"),
        ]),
        WordGroup(id: "verbs", title: "Verbs", note: "Past tense ends in -a; present starts with ya-.", style: .everyday, items: [
            W("kataba", "he wrote"), W("yaktubu", "he writes"), W("qaraAa", "he read"), W("yaqraAu", "he reads"),
            W("dhahaba", "he went"), W("yadhhabu", "he goes"), W("jalasa", "he sat"), W("Aakala", "he ate"),
            W("shariba", "he drank"), W("fahima", "he understood"), W("eamila", "he worked"), W("qaAla", "he said"),
            W("yaquwlu", "he says"), W("kaAna", "he was"), W("jaAAa", "he came"), W("raAaY", "he saw"),
            W("samiea", "he heard"), W("naZara", "he looked"), W("earafa", "he knew"), W("saAala", "he asked"),
            W("dakhala", "he entered"), W("kharaja", "he went out"), W("rajaea", "he returned"), W("fataHa", "he opened"),
            W("taeallama", "he learned"), W("eallama", "he taught"), W("Aaslama", "he submitted to God"),
            W("istagfara", "he asked forgiveness"), W("ijtamaea", "he met / gathered"), W("uktub", "write! (order)"),
        ]),
        WordGroup(id: "little", title: "Little words", note: "Prepositions, pronouns, question words.", style: .everyday, items: [
            W("fiy", "in"), W("min", "from"), W("AilaY", "to"), W("ealaY", "on"), W("ean", "about"),
            W("maea", "with"), W("huwa", "he"), W("hiya", "she"), W("AanaA", "I"), W("Aanta", "you (m.)"),
            W("Aanti", "you (f.)"), W("naHnu", "we"), W("hum", "they"), W("haAdhaA", "this (m.)"), W("haAdhihi", "this (f.)"),
            W("dhaAlika", "that"), W("alladhiy", "who, which (m.)"), W("allatiy", "who, which (f.)"), W("maA", "what"),
            W("man", "who?"), W("hal", "question word"), W("laA", "no"), W("naeam", "yes"), W("Ainna", "indeed"),
            W("Aanna", "that (conj.)"), W("kayfa", "how?"), W("Aayna", "where?"), W("mataY", "when?"),
            W("limaAdhaA", "why?"), W("thumma", "then"), W("qad", "already"), W("laAkin", "but"),
            W("kullu", "all, every"), W("baeDu", "some"), W("lam", "did not"), W("lan", "will not"), W("Aaw", "or"),
        ]),
        WordGroup(id: "hamza", title: "Hamza", note: "A + haraka, or ' when the hamza has no haraka. The seat is chosen for you.", style: .everyday, items: [
            W("saAala", "he asked"), W("suAaAluN", "a question"), W("masAalat'uN", "a matter"), W("yasAalu", "he asks"),
            W("ra'suN", "a head"), W("bi'ruN", "a well"), W("mu'minuN", "a believer"), W("raAiysuN", "a leader"),
            W("qaraAa", "he read"), W("shay'uN", "a thing"), W("shay'aN", "something (object)"), W("juz'uN", "a part"),
            W("samaaAuN", "sky"), W("tasaaAala", "he wondered"), W("lu'luAuN", "pearls"), W("AAmana", "he believed"),
            W("qurAAnuN", "the Qur'an"), W("AiymaAnuN", "faith"), W("Aislaamu", "Islam"), W("Aummat'uN", "a nation"),
            W("mas'uwluN", "responsible"), W("alAarDu", "the earth"), W("alAinsaAnu", "the human being"),
        ]),
        WordGroup(id: "ayn", title: "ʿAyn", note: "ʿAyn is e, as in Eid. Then its haraka: ea, ei, eu.", style: .everyday, items: [
            W("eilmuN", "knowledge"), W("eabduN", "a servant"), W("eaqluN", "mind"), W("eiyduN", "Eid, festival"),
            W("saeiyduN", "happy"), W("maea", "with"), W("naeam", "yes"), W("baeda", "after"),
            W("rabiyeuN", "spring"), W("jamiyeuN", "all"), W("sabeat'uN", "seven"), W("Aarbaeat'uN", "four"),
            W("saAeat'uN", "an hour"), W("mamnuweuN", "forbidden"), W("dueaAAuN", "a supplication"), W("shaeruN", "hair"),
            W("saeala", "he coughed"),
        ]),
        WordGroup(id: "article", title: "The article al-", note: "Sun letters take shadda; moon letters keep the lām with sukūn.", style: .everyday, items: [
            W("alkitaAbu", "the book"), W("alqalamu", "the pen"), W("albaytu", "the house"), W("alqamaru", "the moon"),
            W("alshamsu", "the sun"), W("ash-shamsu", "the sun (typed as heard)"), W("alnuwru", "the light"),
            W("alrajulu", "the man"), W("aldarsu", "the lesson"), W("allaylu", "the night"), W("aleilmu", "the knowledge"),
            W("wal-qamari", "and the moon"), W("wal-shamsi", "and the sun"), W("bil-qalami", "with the pen"),
            W("lil-naAsi", "for the people"),
        ]),
        WordGroup(id: "sentences", title: "Sentences", note: "Space finishes each word; punctuation becomes Arabic.", style: .everyday, items: [
            W("dhahaba alwaladu AilaY almadrasat'i.", "The boy went to the school."),
            W("AanaA AuHibbu allugat'a alearabiyyat'a", "I love the Arabic language."),
            W("hal fahimta?", "Did you understand?"),
            W("haAdhaA kitaAbuN jamiyluN", "This is a beautiful book."),
            W("maA ismuka?", "What is your name?"),
            W("Aayna albaytu?", "Where is the house?"),
            W("alkitaAbu ealaY almaktabi", "The book is on the desk."),
            W("Talaba alTaAlibu alqalama", "The student asked for the pen."),
            W("kaAna aljawwu jamiylaN", "The weather was nice."),
            W("bismi allaAhi", "In the name of Allah"),
        ]),
        WordGroup(id: "fatiha", title: "Qur'an: al-Fātiḥa", note: "Qur'an style: alif waṣla ٱ, dagger alif ^, maddah =.", style: .quran, items: [
            W("\\bism", "1. In the name of Allah, the Most Merciful, the Ever Merciful"),
            W("alHamdu lillaAhi rabbi alea^lamiyna", "2. All praise is for Allah, Lord of the worlds"),
            W("alraHmaAni alraHiymi", "3. the Most Merciful, the Ever Merciful"),
            W("ma^liki yawmi aldiyni", "4. Master of the Day of Judgement"),
            W("Aiyyaaka naebudu wa-Aiyyaaka nastaeiynu", "5. You alone we worship, You alone we ask for help"),
            W("ihdinaa alSira^Ta almustaqiyma", "6. Guide us on the straight path"),
            W("Sira^Ta alladhiyna Aaneamta ealayhim", "7. the path of those You have blessed,"),
            W("gayri almagDuwbi ealayhim wa-laA alDaA=lliyna", "not of those who earned anger, nor of those astray"),
        ]),
        WordGroup(id: "ikhlas", title: "Qur'an: al-Ikhlāṣ", note: "Qur'an style.", style: .quran, items: [
            W("qul huwa allaAhu AaHaduN", "1. Say: He is Allah, One"),
            W("allaAhu alSamadu", "2. Allah, the Self-Sufficient"),
            W("lam yalid wa-lam yuwlad", "3. He does not beget, nor was He begotten"),
            W("wa-lam yakun lahu\\w kufuwaN AaHaduN", "4. and there is none comparable to Him"),
        ]),
    ]

    /// A short mixed selection shown under the letter table.
    public static let featured: [WordItem] = {
        let pick: [(String, Int)] = [("first", 8), ("verbs", 8), ("little", 8), ("hamza", 4), ("ayn", 4), ("article", 4)]
        var out: [WordItem] = []
        for (id, n) in pick {
            if let g = groups.first(where: { $0.id == id }) { out += g.items.prefix(n) }
        }
        return out
    }()

    // MARK: Other ways to type the same thing

    /// Other spellings that give exactly the same Arabic:
    /// long vowels doubled (`eaA` → `eaa`, `eiy` → `eii`, `euw` → `euu`) and hamza with `'` (`saAala` → `sa'ala`).
    public static func alsoTypes(_ latin: String) -> [String] {
        let c = Array(latin)
        func at(_ i: Int) -> Character? { i >= 0 && i < c.count ? c[i] : nil }

        var doubled = ""
        for (i, ch) in c.enumerated() {
            let prev = at(i - 1), next = at(i + 1)
            if ch == "A", prev == "a", !(next.map { "aiuNA".contains($0) } ?? false) {
                doubled.append("a")
            } else if ch == "y", prev == "i", !(next.map { "aiuyN".contains($0) } ?? false) {
                doubled.append("i")
            } else if ch == "w", prev == "u", !(next.map { "aiuwN".contains($0) } ?? false) {
                doubled.append("u")
            } else {
                doubled.append(ch)
            }
        }

        var apostrophe = ""
        for (i, ch) in c.enumerated() {
            let next = at(i + 1)
            apostrophe.append(ch == "A" && (next.map { "aiu".contains($0) } ?? false) ? "'" : ch)
        }

        var out: [String] = []
        for v in [doubled, apostrophe] where v != latin && !out.contains(v) {
            let same = Style.allCases.allSatisfy { style in
                var o = Options()
                o.style = style
                return Qalam.text(v, o) == Qalam.text(latin, o)
            }
            if same { out.append(v) }
        }
        return out
    }

    // MARK: Drills

    /// Real words for one letter, or for every letter.
    public static func letterDrills(for arabic: String? = nil) -> [PracticeItem] {
        letters.filter { arabic == nil || $0.arabic == arabic }.flatMap { row in
            row.drills.map { PracticeItem(latin: $0.latin, label: "\(row.arabic)  \(row.name)", meaning: $0.meaning, style: .everyday) }
        }
    }

    public static func items(ofGroup id: String) -> [PracticeItem] {
        guard let g = groups.first(where: { $0.id == id }) else { return [] }
        return g.items.map { PracticeItem(latin: $0.latin, label: "", meaning: $0.meaning, style: g.style) }
    }
}
