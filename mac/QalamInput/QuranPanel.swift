import Cocoa
import QalamEngine

/// The Qur'an data inside Qalam.app (Contents/Resources/Quran), plus translations the app downloaded.
enum QuranStore {
    static let shared: QuranLibrary? = {
        var dirs: [URL] = []
        if let app = NSWorkspace.shared.urlForApplication(withBundleIdentifier: "com.qalam.app") {
            dirs.append(app.appendingPathComponent("Contents/Resources/Quran"))
        }
        dirs.append(FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Applications/Qalam.app/Contents/Resources/Quran"))
        return QuranLibrary(folders: dirs)
    }()
}

/// What the Qur'an box shows: the Choose tab (Sūrah / From / To dropdowns) or the Search tab.
struct QuranPick {
    enum Tab { case choose, search }
    enum Field: Int, CaseIterable { case surah, from, to }

    var tab = Tab.choose
    var surah = 1
    var from = 1
    var to = 1
    var field = Field.surah
    /// Choose tab: what's typed into the highlighted dropdown ("baq", "25"). Search tab: the query.
    var typed = ""
    var query = ""
    var results: [QuranSuggestion] = []
    var selected = 0
    var total: Int?

    var passage: Passage { Passage(surah: surah, from: from, to: max(from, to)) }

    /// Opens where it was last time (stored as "2:255-257").
    static func last(_ lib: QuranLibrary) -> QuranPick {
        var p = QuranPick()
        if let saved = SharedSettings.quranLast, let r = lib.resolve(saved) {
            p.surah = r.surah; p.from = r.from; p.to = r.to
        }
        return p
    }
}

/// The "insert from the Qur'an" box (⌃⇧Q). Like the options panel it never takes focus: the keyboard
/// sends it the keys, and its dropdowns and buttons work with the mouse.
final class QuranPanel {
    static let shared = QuranPanel()

    var onSurah: ((Int) -> Void)?
    var onFrom: ((Int) -> Void)?
    var onTo: ((Int) -> Void)?
    var onField: ((QuranPick.Field) -> Void)?
    var onTab: ((QuranPick.Tab) -> Void)?
    var onPickResult: ((Int) -> Void)?
    var onMode: ((QuranInsert) -> Void)?
    var onTranslation: ((String) -> Void)?
    var onInsert: (() -> Void)?
    var onClose: (() -> Void)?

    private var panel: NSPanel?
    private let width: CGFloat = 580
    private var anchor = NSRect.zero

    private lazy var quranFont: NSFont = {
        OptionsPanel.registerQalamFonts()
        for family in ["KFGQPC HAFS Uthmanic Script", "KFGQPC Uthmanic Script HAFS", "Scheherazade New", "Amiri Quran"] {
            if let f = NSFontManager.shared.font(withFamily: family, traits: [], weight: 5, size: 22) { return f }
        }
        return NSFont.systemFont(ofSize: 22)
    }()

    var snapshotView: NSView? { panel?.contentView }
    var isVisible: Bool { panel?.isVisible ?? false }

    func hide() { panel?.orderOut(nil) }

    /// `caret`: where the text cursor is; the box keeps its first position while it stays open.
    func show(_ p: QuranPick, library lib: QuranLibrary, mode: QuranInsert, translation: String,
              near caret: NSRect, key: String?) {
        let panel = self.panel ?? OptionsPanel.makePanel()
        if self.panel == nil || !panel.isVisible { anchor = caret }
        self.panel = panel

        var views: [NSView] = [header(p)]
        switch p.tab {
        case .choose: views += chooseViews(p, lib, mode: mode, translation: translation)
        case .search: views += searchViews(p, lib)
        }
        views += [separator(), insertRow(p, lib, mode: mode, translation: translation), hint(p, key: key)]

        let stack = NSStackView(views: views)
        stack.orientation = .vertical
        stack.alignment = .leading
        stack.spacing = 8
        stack.edgeInsets = NSEdgeInsets(top: 10, left: 10, bottom: 10, right: 10)
        for v in views { v.widthAnchor.constraint(equalToConstant: width - 20).isActive = true }

        let background = NSVisualEffectView()
        background.material = .popover
        background.state = .active
        background.wantsLayer = true
        background.layer?.cornerRadius = 10
        background.layer?.masksToBounds = true
        background.addSubview(stack)
        stack.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            stack.leadingAnchor.constraint(equalTo: background.leadingAnchor),
            stack.trailingAnchor.constraint(equalTo: background.trailingAnchor),
            stack.topAnchor.constraint(equalTo: background.topAnchor),
            stack.bottomAnchor.constraint(equalTo: background.bottomAnchor),
        ])
        let size = NSSize(width: width, height: stack.fittingSize.height)
        panel.contentView = background
        panel.setContentSize(size)
        panel.setFrameOrigin(OptionsPanel.origin(for: size, near: anchor))
        panel.orderFrontRegardless()
    }

    // MARK: Header: title and the two tabs

    private func header(_ p: QuranPick) -> NSView {
        let title = NSTextField(labelWithString: "Insert from the Qur'an")
        title.font = .systemFont(ofSize: 13, weight: .semibold)
        let tabs = NSStackView(views: [
            chip("Choose", on: p.tab == .choose) { [weak self] in self?.onTab?(.choose) },
            chip("Search", on: p.tab == .search) { [weak self] in self?.onTab?(.search) },
        ])
        tabs.spacing = 4
        let close = chip("✕", on: false) { [weak self] in self?.onClose?() }
        close.toolTip = "Close (Esc)"
        let h = NSStackView(views: [title, NSView(), tabs, close])
        h.orientation = .horizontal
        return h
    }

    // MARK: Choose: Sūrah ▾  From ▾  To ▾, then a preview

    private func chooseViews(_ p: QuranPick, _ lib: QuranLibrary, mode: QuranInsert, translation: String) -> [NSView] {
        guard let s = lib.surah(p.surah) else { return [] }
        let surahMenu = popup(lib.surahs.map { "\($0.number). \($0.name)  ·  \($0.english)  ·  \($0.ayahs) āyāt  ·  \($0.arabic)\u{200E}" },
                              selected: p.surah - 1) { [weak self] i in self?.onSurah?(i + 1) }
        let fromMenu = popup((1...s.ayahs).map(String.init), selected: p.from - 1) { [weak self] i in self?.onFrom?(i + 1) }
        let toMenu = popup((p.from...s.ayahs).map(String.init), selected: max(p.from, p.to) - p.from) { [weak self] i in
            self?.onTo?(p.from + i)
        }
        let row = NSStackView(views: [
            field("Sūrah", surahMenu, typed: p.field == .surah ? p.typed : "", on: p.field == .surah, .surah),
            field("From āyah", fromMenu, typed: p.field == .from ? p.typed : "", on: p.field == .from, .from),
            field("To āyah", toMenu, typed: p.field == .to ? p.typed : "", on: p.field == .to, .to),
        ])
        row.orientation = .horizontal
        row.alignment = .top
        row.spacing = 8
        row.distribution = .fill
        surahMenu.widthAnchor.constraint(equalToConstant: 320).isActive = true
        fromMenu.widthAnchor.constraint(equalToConstant: 90).isActive = true
        toMenu.widthAnchor.constraint(equalToConstant: 90).isActive = true

        return [row, preview(p.passage, lib, mode: mode, translation: translation)]
    }

    /// The first āyāt of the choice, in the Qur'an font, with the translation under the first one.
    private func preview(_ passage: Passage, _ lib: QuranLibrary, mode: QuranInsert, translation: String) -> NSView {
        var parts: [NSView] = []
        let count = passage.to - passage.from + 1
        let name = lib.surah(passage.surah)?.name ?? ""
        let ref = NSTextField(labelWithString: "\(name) \(passage.reference)" + (count > 1 ? "  ·  \(count) āyāt" : ""))
        ref.font = .systemFont(ofSize: 11, weight: .medium)
        ref.textColor = .secondaryLabelColor
        parts.append(ref)
        let shownAyat = passage.from...min(passage.to, passage.from + 2)
        let arabic = shownAyat.map { "\(lib.arabic(passage.surah, $0) ?? "") \(arabicNumber($0))" }.joined(separator: " ")
        parts.append(arabicLabel(arabic, lines: 4))
        if count > shownAyat.count { parts.append(small("… and \(count - shownAyat.count) more āyāt")) }
        if mode != .arabic, let tr = lib.translation(translation, passage.surah, passage.from) {
            let l = NSTextField(wrappingLabelWithString: tr)
            l.font = .systemFont(ofSize: 12)
            l.textColor = .secondaryLabelColor
            l.maximumNumberOfLines = 3
            l.lineBreakMode = .byTruncatingTail
            l.preferredMaxLayoutWidth = width - 44
            parts.append(l)
        }
        return card(parts, selected: false, action: nil)
    }

    // MARK: Search: a field and the results

    private func searchViews(_ p: QuranPick, _ lib: QuranLibrary) -> [NSView] {
        let field: NSTextField
        if p.query.isEmpty {
            field = NSTextField(labelWithString: "Type a word: الرحمن, ٱلرَّحْمَٰنِ or rahmaan (harakat optional)")
            field.textColor = .placeholderTextColor
        } else {
            field = NSTextField(labelWithString: p.query + "▏")
        }
        field.font = .systemFont(ofSize: 14)
        var views: [NSView] = [box(field)]
        for (i, r) in p.results.enumerated() {
            switch r {
            case .hit(let x, _), .passage(let x):
                let ref = NSTextField(labelWithString: "\(lib.surah(x.surah)?.name ?? "") \(x.reference)")
                ref.font = .systemFont(ofSize: 11, weight: .medium)
                ref.textColor = .secondaryLabelColor
                views.append(card([ref, arabicLabel(lib.arabic(x.surah, x.from) ?? "", lines: i == p.selected ? 3 : 1)],
                                  selected: i == p.selected) { [weak self] in self?.onPickResult?(i) })
            case .surah(let s):
                let l = NSTextField(labelWithString: "\(s.number). \(s.name)  ·  \(s.english)  ·  \(s.arabic)")
                l.font = .systemFont(ofSize: 13)
                views.append(card([l], selected: i == p.selected) { [weak self] in self?.onPickResult?(i) })
            case .problem(let text):
                views.append(small(text))
            }
        }
        if let total = p.total, total > p.results.count {
            views.append(small("\(total) āyāt found, showing \(p.results.count). Type more to narrow it down."))
        }
        return views
    }

    // MARK: Bottom: what to insert, and the Insert button

    private func insertRow(_ p: QuranPick, _ lib: QuranLibrary, mode: QuranInsert, translation: String) -> NSView {
        let modes = NSStackView(views: QuranInsert.allCases.map { m in
            chip(m.title, on: m == mode) { [weak self] in self?.onMode?(m) }
        })
        modes.spacing = 4
        var items: [NSView] = [modes]
        if mode != .arabic {
            let have = lib.available()
            let menu = popup(have.map(\.title), selected: have.firstIndex { $0.id == translation } ?? 0) { [weak self] i in
                self?.onTranslation?(have[i].id)
            }
            menu.controlSize = .small
            menu.font = .systemFont(ofSize: 11)
            items.append(menu)
        }
        items.append(NSView())
        let canInsert = p.tab == .choose || p.results.contains { if case .problem = $0 { return false }; return true }
        let insert = chip("  Insert  ⏎  ", on: canInsert) { [weak self] in if canInsert { self?.onInsert?() } }
        insert.toolTip = "Insert (Return)"
        items.append(insert)
        let h = NSStackView(views: items)
        h.orientation = .horizontal
        h.spacing = 8
        return h
    }

    private func hint(_ p: QuranPick, key: String?) -> NSView {
        switch p.tab {
        case .choose:
            return small("↑↓ change · ←→ or Tab next box · type 36 or yasin to jump, 2:255-257 to fill all · Return insert · / search · Esc close")
        case .search:
            return small("↑↓ choose · Return insert · Tab back to Choose · Esc close" + (key.map { " · \($0) opens this box" } ?? ""))
        }
    }

    // MARK: Pieces

    /// A labelled dropdown; the highlighted one (keyboard focus) gets an accent border and shows what's typed.
    private func field(_ label: String, _ control: NSView, typed: String, on: Bool, _ f: QuranPick.Field) -> NSView {
        let l = NSTextField(labelWithString: typed.isEmpty ? label : "\(label): \(typed)▏")
        l.font = .systemFont(ofSize: 11, weight: on ? .semibold : .regular)
        l.textColor = on ? .controlAccentColor : .secondaryLabelColor
        let v = NSStackView(views: [l, control])
        v.orientation = .vertical
        v.alignment = .leading
        v.spacing = 3
        let click = ClickView { [weak self] in self?.onField?(f) }
        click.wantsLayer = true
        click.layer?.cornerRadius = 7
        click.layer?.borderWidth = on ? 2 : 0
        click.layer?.borderColor = NSColor.controlAccentColor.cgColor
        click.addSubview(v)
        v.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            v.leadingAnchor.constraint(equalTo: click.leadingAnchor, constant: 5),
            v.trailingAnchor.constraint(equalTo: click.trailingAnchor, constant: -5),
            v.topAnchor.constraint(equalTo: click.topAnchor, constant: 4),
            v.bottomAnchor.constraint(equalTo: click.bottomAnchor, constant: -5),
        ])
        return click
    }

    private func popup(_ titles: [String], selected: Int, action: @escaping (Int) -> Void) -> NSPopUpButton {
        let b = FirstMousePopUp(frame: .zero, pullsDown: false)
        b.addItems(withTitles: titles)
        if titles.indices.contains(selected) { b.selectItem(at: selected) }
        b.onChange = action
        return b
    }

    private func card(_ parts: [NSView], selected: Bool, action: (() -> Void)?) -> NSView {
        let v = NSStackView(views: parts)
        v.orientation = .vertical
        v.alignment = .leading
        v.spacing = 3
        for p in parts { p.widthAnchor.constraint(equalToConstant: width - 44).isActive = true }
        let click = ClickView { action?() }
        click.wantsLayer = true
        click.layer?.cornerRadius = 7
        click.layer?.backgroundColor = (selected ? NSColor.controlAccentColor.withAlphaComponent(0.22)
                                                 : NSColor.textBackgroundColor.withAlphaComponent(0.35)).cgColor
        click.addSubview(v)
        v.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            v.leadingAnchor.constraint(equalTo: click.leadingAnchor, constant: 12),
            v.trailingAnchor.constraint(equalTo: click.trailingAnchor, constant: -12),
            v.topAnchor.constraint(equalTo: click.topAnchor, constant: 6),
            v.bottomAnchor.constraint(equalTo: click.bottomAnchor, constant: -6),
        ])
        return click
    }

    private func arabicLabel(_ text: String, lines: Int) -> NSTextField {
        let style = NSMutableParagraphStyle()
        style.alignment = .right
        style.baseWritingDirection = .rightToLeft
        style.lineBreakMode = .byWordWrapping
        let l = NSTextField(wrappingLabelWithString: "")
        let shown = quranFont.familyName?.contains("KFGQPC") == true ? QuranLibrary.forKFGQPC(text) : text
        l.attributedStringValue = NSAttributedString(string: shown, attributes: [.font: quranFont, .paragraphStyle: style])
        l.maximumNumberOfLines = lines
        l.cell?.truncatesLastVisibleLine = true
        l.preferredMaxLayoutWidth = width - 44
        return l
    }

    /// The āyah-end mark with its number, as a mushaf has it. KFGQPC Hafs circles plain Arabic digits by
    /// itself (۝ + digits would draw two circles); other fonts need ۝ (U+06DD) before the digits.
    private func arabicNumber(_ n: Int) -> String {
        let digits = String(String(n).map { Character(Unicode.Scalar(0x0660 + Int(String($0))!)!) })
        return quranFont.familyName?.contains("KFGQPC") == true ? digits : "\u{06DD}" + digits
    }

    private func chip(_ title: String, on: Bool, action: @escaping () -> Void) -> NSView {
        let label = NSTextField(labelWithString: title)
        label.font = .systemFont(ofSize: 11, weight: on ? .semibold : .regular)
        label.textColor = on ? .white : .labelColor
        let v = ClickView(action: action)
        v.wantsLayer = true
        v.layer?.cornerRadius = 5
        v.layer?.borderWidth = on ? 0 : 1
        v.layer?.borderColor = NSColor.separatorColor.cgColor
        v.layer?.backgroundColor = on ? NSColor.controlAccentColor.cgColor : NSColor.clear.cgColor
        v.addSubview(label)
        label.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            label.leadingAnchor.constraint(equalTo: v.leadingAnchor, constant: 8),
            label.trailingAnchor.constraint(equalTo: v.trailingAnchor, constant: -8),
            label.topAnchor.constraint(equalTo: v.topAnchor, constant: 3),
            label.bottomAnchor.constraint(equalTo: v.bottomAnchor, constant: -3),
        ])
        return v
    }

    private func box(_ content: NSView) -> NSView {
        let v = NSView()
        v.wantsLayer = true
        v.layer?.cornerRadius = 6
        v.layer?.borderWidth = 2
        v.layer?.borderColor = NSColor.controlAccentColor.cgColor
        v.layer?.backgroundColor = NSColor.textBackgroundColor.withAlphaComponent(0.6).cgColor
        v.addSubview(content)
        content.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            content.leadingAnchor.constraint(equalTo: v.leadingAnchor, constant: 8),
            content.trailingAnchor.constraint(lessThanOrEqualTo: v.trailingAnchor, constant: -8),
            content.topAnchor.constraint(equalTo: v.topAnchor, constant: 6),
            content.bottomAnchor.constraint(equalTo: v.bottomAnchor, constant: -6),
        ])
        return v
    }

    private func small(_ s: String) -> NSTextField {
        let l = NSTextField(wrappingLabelWithString: s)
        l.font = .systemFont(ofSize: 10)
        l.textColor = .tertiaryLabelColor
        return l
    }

    private func separator() -> NSView {
        let b = NSBox()
        b.boxType = .separator
        return b
    }
}

/// A dropdown that opens on the first click even though its window is never the key window.
final class FirstMousePopUp: NSPopUpButton {
    var onChange: ((Int) -> Void)?
    override init(frame: NSRect, pullsDown: Bool) {
        super.init(frame: frame, pullsDown: pullsDown)
        target = self
        action = #selector(changed)
    }
    required init?(coder: NSCoder) { fatalError() }
    override func acceptsFirstMouse(for event: NSEvent?) -> Bool { true }
    @objc private func changed() { onChange?(indexOfSelectedItem) }
}
