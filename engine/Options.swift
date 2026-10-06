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

public struct Options: Equatable, Sendable {
    public var style: Style
    public var harakat: HarakatMode
    /// In Qur'an style, write sukūn as ۡ (Madinah mushaf) instead of ْ.
    public var quranSmallSukun: Bool
    /// Spell words like اللَّه، هَٰذَا، ذَٰلِكَ the way they are written, not as they sound.
    public var spellingWords: Bool
    public var arabicDigits: Bool

    public init(style: Style = .everyday, harakat: HarakatMode = .full, quranSmallSukun: Bool = false,
                spellingWords: Bool = true, arabicDigits: Bool = false) {
        self.style = style
        self.harakat = harakat
        self.quranSmallSukun = quranSmallSukun
        self.spellingWords = spellingWords
        self.arabicDigits = arabicDigits
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
        if d.object(forKey: "quranSmallSukun") != nil { o.quranSmallSukun = d.bool(forKey: "quranSmallSukun") }
        if d.object(forKey: "spellingWords") != nil { o.spellingWords = d.bool(forKey: "spellingWords") }
        if d.object(forKey: "arabicDigits") != nil { o.arabicDigits = d.bool(forKey: "arabicDigits") }
        return o
    }

    public static func save(_ o: Options) {
        let d = defaults
        d.set(o.style.rawValue, forKey: "style")
        d.set(o.harakat.rawValue, forKey: "harakat")
        d.set(o.quranSmallSukun, forKey: "quranSmallSukun")
        d.set(o.spellingWords, forKey: "spellingWords")
        d.set(o.arabicDigits, forKey: "arabicDigits")
        d.synchronize()
    }
}
