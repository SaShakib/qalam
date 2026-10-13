import SwiftUI
import AppKit
import QalamEngine

enum Page: String, CaseIterable, Identifiable, Hashable {
    case welcome, letters, words, practice, write, settings
    var id: String { rawValue }

    var title: String {
        switch self {
        case .welcome: return "Welcome"
        case .letters: return "Letters & keys"
        case .words: return "Word list"
        case .practice: return "Practice"
        case .write: return "Write"
        case .settings: return "Settings"
        }
    }

    var icon: String {
        switch self {
        case .welcome: return "hand.wave"
        case .letters: return "character.book.closed"
        case .words: return "list.bullet.rectangle"
        case .practice: return "keyboard"
        case .write: return "square.and.pencil"
        case .settings: return "gearshape"
        }
    }
}

enum PracticeSet: Hashable {
    case starter
    case letters
    case letter(String)
    case group(String)
}

final class AppState: ObservableObject {
    @Published var page: Page?
    @Published var practiceSet: PracticeSet = .starter
    @Published var firstPracticeDone: Bool
    /// The setup steps were finished or skipped (people who used Qalam before 1.3 count as done).
    @Published var setupDone: Bool
    @Published var options: Options {
        didSet { SharedSettings.save(options); shown.removeAll() }
    }
    /// Arabic already produced for (latin, style): pages redraw often, so don't redo the work.
    private var shown: [String: String] = [:]
    @Published var installedFonts: [String] = []
    @Published var kfgqpcFamily: String?
    @Published var fontName: String {
        didSet { UserDefaults.standard.set(fontName, forKey: "arabicFont3") }
    }
    @Published var quranFontName: String {
        didSet { UserDefaults.standard.set(quranFontName, forKey: "quranFont2") }
    }
    @Published var fontScale: Double {
        didSet { UserDefaults.standard.set(fontScale, forKey: "fontScale") }
    }

    init() {
        options = SharedSettings.load()
        ArabicFonts.registerBundled()
        // A font the user picked stays; otherwise SIL Scheherazade New, and KFGQPC Hafs for Qur'an text once it's there.
        fontName = UserDefaults.standard.string(forKey: "arabicFont3") ?? ArabicFonts.defaultArabic
        quranFontName = UserDefaults.standard.string(forKey: "quranFont2") ?? ArabicFonts.defaultQuran()
        let scale = UserDefaults.standard.double(forKey: "fontScale")
        fontScale = scale == 0 ? 1 : scale
        let practiced = UserDefaults.standard.bool(forKey: "firstPracticeDone")
        firstPracticeDone = practiced
        setupDone = UserDefaults.standard.object(forKey: "setupDone") as? Bool ?? practiced
        page = .letters
    }

    func finishFirstPractice() {
        UserDefaults.standard.set(true, forKey: "firstPracticeDone")
        firstPracticeDone = true
        page = .letters
    }

    /// End of the setup steps: on to the 20-word practice, or straight into the app.
    func finishSetup(practice: Bool) {
        UserDefaults.standard.set(true, forKey: "setupDone")
        setupDone = true
        if practice {
            restartFirstPractice()
        } else if !firstPracticeDone {
            finishFirstPractice()
        }
    }

    func restartSetup() {
        UserDefaults.standard.set(false, forKey: "setupDone")
        setupDone = false
    }

    func restartFirstPractice() {
        UserDefaults.standard.set(false, forKey: "firstPracticeDone")
        firstPracticeDone = false
    }

    /// The Arabic display font (the Qur'an font for Qur'an-style text).
    func arabic(_ size: CGFloat, quran: Bool = false) -> Font {
        let s = size * fontScale
        let name = quran ? quranFontName : fontName
        return name.isEmpty ? .system(size: s) : .custom(name, size: s)
    }

    /// What the keyboard produces for `latin`, always with full harakat (for learning).
    func show(_ latin: String, style: Style? = nil) -> String {
        var o = options
        o.harakat = .full
        if let style { o.style = style }
        let key = "\(o.style.rawValue)|\(latin)"
        if let s = shown[key] { return s }
        let s = Qalam.text(latin, o)
        shown[key] = s
        return s
    }

    /// What the keyboard writes for `latin` with the current settings changed by `change`
    /// (the setup steps show each choice before it is picked).
    func preview(_ latin: String, _ change: (inout Options) -> Void) -> String {
        var o = options
        change(&o)
        return Qalam.text(latin, o)
    }

    /// Other ways to type `latin` (cached).
    func also(_ latin: String) -> [String] {
        if let a = alsoCache[latin] { return a }
        let a = Content.alsoTypes(latin)
        alsoCache[latin] = a
        return a
    }
    private var alsoCache: [String: [String]] = [:]

    /// Font scan runs in the background (it takes ~1 s on a Mac with many fonts).
    func scanFonts() {
        guard installedFonts.isEmpty else { return }
        DispatchQueue.global(qos: .utility).async {
            let (list, kfgqpc) = ArabicFonts.scanInstalled()
            DispatchQueue.main.async {
                self.installedFonts = list
                self.kfgqpcFamily = kfgqpc
            }
        }
    }

    func practise(_ set: PracticeSet) {
        practiceSet = set
        page = .practice
    }

    var practiceChoices: [(PracticeSet, String)] {
        var c: [(PracticeSet, String)] = [(.starter, "Starter: 20 key words"), (.letters, "All letters (real words)")]
        if case .letter(let l) = practiceSet,
           let row = Content.letters.first(where: { $0.arabic == l }) {
            c.append((practiceSet, "One letter: \(l)  \(row.name)"))
        }
        c += Content.groups.map { (.group($0.id), "Words: \($0.title)") }
        return c
    }

    func items(_ set: PracticeSet) -> [PracticeItem] {
        switch set {
        case .starter: return Content.starter
        case .letters: return Content.letterDrills()
        case .letter(let l): return Content.letterDrills(for: l)
        case .group(let id): return Content.items(ofGroup: id)
        }
    }


}

/// A key the user presses, drawn like a keycap.
struct KeyCap: View {
    let text: String
    init(_ text: String) { self.text = text }
    var body: some View {
        Text(text)
            .font(.system(.body, design: .monospaced).weight(.semibold))
            .padding(.horizontal, 7)
            .padding(.vertical, 2)
            .background(RoundedRectangle(cornerRadius: 5).fill(Color.accentColor.opacity(0.14)))
            .overlay(RoundedRectangle(cornerRadius: 5).stroke(Color.accentColor.opacity(0.35), lineWidth: 1))
    }
}

struct PageHeader: View {
    let title: String
    let subtitle: String
    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(title).font(.largeTitle.bold())
            Text(subtitle).foregroundStyle(.secondary)
        }
    }
}

struct SectionTitle: View {
    let text: String
    init(_ text: String) { self.text = text }
    var body: some View {
        Text(text).font(.title2.bold()).padding(.top, 8)
    }
}
