import Foundation

public enum Style: String, CaseIterable, Identifiable, Sendable {
    case everyday, quran
    public var id: String { rawValue }
    public var title: String { self == .everyday ? "Everyday" : "Qur'an (Uthmani)" }
}

public enum HarakatMode: String, CaseIterable, Identifiable, Sendable {
    case full, asTyped, none
    public var id: String { rawValue }
    public var title: String {
        switch self {
        case .full: return "Full (automatic sukūn)"
        case .asTyped: return "Only what I type"
        case .none: return "None (plain text)"
        }
    }
}

/// How sukūn is added (only when harakat is Full).
public enum SukunMode: String, CaseIterable, Identifiable, Sendable {
    /// Only at a real stop inside a word (vowel before, consonant after); `o` always adds one.
    case smart
    /// On every vowel-less letter (Qur'an style always works this way unless Off).
    case full
    /// None at all, not even from `o`.
    case off
    public var id: String { rawValue }
    public var title: String {
        switch self {
        case .smart: return "Smart (only where needed)"
        case .full: return "Full (every stop)"
        case .off: return "Off (no sukūn)"
        }
    }
}

public struct Options: Equatable, Sendable {
    public var style: Style
    public var harakat: HarakatMode
    public var sukun: SukunMode
    /// In Qur'an style, write sukūn as ۡ (Madinah mushaf) instead of ْ.
    public var quranSmallSukun: Bool
    /// Spell words like اللَّه، هَٰذَا، ذَٰلِكَ the way they are written, not as they sound.
    public var spellingWords: Bool
    public var arabicDigits: Bool
    /// Chrome-based apps (HarfBuzz) reorder harakat and then draw the font's built-in "Allah"
    /// ligature on top of our marks (doubled shadda/alif). When true, an invisible zero-width joiner
    /// (U+200D) between the two lāms of لله stops that ligature. Set automatically by the keyboard.
    public var blockAllahLigature: Bool

    public init(style: Style = .everyday, harakat: HarakatMode = .full, sukun: SukunMode = .smart, quranSmallSukun: Bool = false,
                spellingWords: Bool = true, arabicDigits: Bool = false, blockAllahLigature: Bool = false) {
        self.style = style
        self.harakat = harakat
        self.sukun = sukun
        self.quranSmallSukun = quranSmallSukun
        self.spellingWords = spellingWords
        self.arabicDigits = arabicDigits
        self.blockAllahLigature = blockAllahLigature
    }
}

/// Settings shared by the keyboard and the app (~/Library/Preferences/com.qalam.shared.plist).
public enum SharedSettings {
    public static let suiteName = "com.qalam.shared"
    static let defaults = UserDefaults(suiteName: suiteName) ?? .standard

    public static func load() -> Options {
        let d = defaults
        var o = Options()
        if let s = d.string(forKey: "style"), let v = Style(rawValue: s) { o.style = v }
        if let s = d.string(forKey: "harakat"), let v = HarakatMode(rawValue: s) { o.harakat = v }
        if let s = d.string(forKey: "sukun"), let v = SukunMode(rawValue: s) { o.sukun = v }
        if d.object(forKey: "quranSmallSukun") != nil { o.quranSmallSukun = d.bool(forKey: "quranSmallSukun") }
        if d.object(forKey: "spellingWords") != nil { o.spellingWords = d.bool(forKey: "spellingWords") }
        if d.object(forKey: "arabicDigits") != nil { o.arabicDigits = d.bool(forKey: "arabicDigits") }
        return o
    }

    /// Set once the user has turned the Qalam keyboard on, so reinstalls and updates can turn it back on.
    public static var keyboardWasEnabled: Bool {
        get { defaults.bool(forKey: "keyboardWasEnabled") }
        set { defaults.set(newValue, forKey: "keyboardWasEnabled") }
    }

    /// Show the options panel while typing (default on).
    public static var showOptions: Bool {
        get { defaults.object(forKey: "showOptions") as? Bool ?? true }
        set { defaults.set(newValue, forKey: "showOptions") }
    }

    // MARK: What the user picked before (stored only on this Mac)

    /// Typed word → the Arabic the user chose for it last time.
    public static func learnedChoice(for typed: String) -> String? {
        (defaults.dictionary(forKey: "learnedChoices") as? [String: String])?[typed]
    }

    /// The user's harakat habit: true after they picked a version without harakat.
    public static var prefersPlain: Bool {
        get { defaults.bool(forKey: "prefersPlain") }
        set { defaults.set(newValue, forKey: "prefersPlain") }
    }

    public static func remember(_ chosen: Candidate, for typed: String, among all: [Candidate]) {
        var map = (defaults.dictionary(forKey: "learnedChoices") as? [String: String]) ?? [:]
        if map.count > 5000 { map.removeAll() }
        map[typed] = chosen.text
        defaults.set(map, forKey: "learnedChoices")
        // Harakat habit: picking the plain version means "no harakat", anything else with marks means "harakat".
        if chosen.kind == .plain {
            prefersPlain = true
        } else if all.contains(where: { $0.kind == .plain && $0.text != chosen.text }) {
            prefersPlain = false
        }
    }

    public static func forgetChoices() {
        defaults.removeObject(forKey: "learnedChoices")
        defaults.removeObject(forKey: "prefersPlain")
    }

    public static func save(_ o: Options) {
        let d = defaults
        d.set(o.style.rawValue, forKey: "style")
        d.set(o.harakat.rawValue, forKey: "harakat")
        d.set(o.sukun.rawValue, forKey: "sukun")
        d.set(o.quranSmallSukun, forKey: "quranSmallSukun")
        d.set(o.spellingWords, forKey: "spellingWords")
        d.set(o.arabicDigits, forKey: "arabicDigits")
        d.synchronize()
    }
}
