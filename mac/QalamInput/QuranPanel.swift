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

/// The "insert from the Qur'an" box (⌃⇧Q): what you type shows in its field, the dropdown under it
/// lists sūrahs, the passage, or search results with the Arabic in the Qur'an font.
/// Like the options panel it never takes focus; the keyboard sends it the keys.
final class QuranPanel {
    static let shared = QuranPanel()

    var onPick: ((Int) -> Void)?
    var onCycleMode: (() -> Void)?
    var onCycleTranslation: (() -> Void)?

    private var panel: NSPanel?
    private let width: CGFloat = 560

    private lazy var quranFont: NSFont = {
        OptionsPanel.registerQalamFonts()
        for family in ["KFGQPC HAFS Uthmanic Script", "KFGQPC Uthmanic Script HAFS", "Scheherazade New", "Amiri Quran"] {
            if let f = NSFontManager.shared.font(withFamily: family, traits: [], weight: 5, size: 22) { return f }
        }
        return NSFont.systemFont(ofSize: 22)
    }()

    var snapshotView: NSView? { panel?.contentView }

    func hide() { panel?.orderOut(nil) }

    func show(query: String, rows: [QuranSuggestion], selected: Int, total: Int?, library lib: QuranLibrary,
              mode: QuranInsert, translation: QuranTranslation?, near caret: NSRect, key: String?) {
        let panel = self.panel ?? OptionsPanel.makePanel()
        self.panel = panel

        // The field: what's typed, or the placeholder.
        let field: NSTextField
        if query.isEmpty {
            field = NSTextField(labelWithString: "2:255 · 2:255-257 · baqarah 255 · or search: rahmaan, الرحمن")
            field.textColor = .placeholderTextColor
        } else {
            field = NSTextField(labelWithString: query + "▏")
            field.textColor = .labelColor
        }
        field.font = .systemFont(ofSize: 14)
        let title = NSTextField(labelWithString: "Qur'an")
        title.font = .systemFont(ofSize: 11, weight: .semibold)
        title.textColor = .secondaryLabelColor
        let fieldBox = box(NSStackView(views: [title, field]), border: true)

        let list = NSStackView()
        list.orientation = .vertical
        list.alignment = .leading
        list.spacing = 2
        for (i, r) in rows.enumerated() {
            list.addArrangedSubview(row(i, r, selected: i == selected, lib: lib, mode: mode, translation: translation))
        }

        var views: [NSView] = [fieldBox]
        if !rows.isEmpty { views.append(list) }
        if let total, total > rows.count {
            views.append(small("\(total) āyāt found, showing \(rows.count). Type more to narrow it down."))
        }
        let modeTitle: String
        switch mode {
        case .arabic: modeTitle = "Insert: Arabic only"
        case .both: modeTitle = "Insert: Arabic + translation"
        case .translation: modeTitle = "Insert: translation only"
        }
        var buttons: [NSView] = [button(modeTitle, key: "Tab") { [weak self] in self?.onCycleMode?() }]
        if mode != .arabic {
            buttons.append(button(translation?.title ?? "No translation", key: "⇧Tab") { [weak self] in self?.onCycleTranslation?() })
        }
        let buttonRow = NSStackView(views: buttons)
        buttonRow.spacing = 6
        views += [separator(), buttonRow,
                  small("↑↓ choose · Return insert · Esc close" + (key.map { " · \($0) opens this box" } ?? ""))]

        let stack = NSStackView(views: views)
        stack.orientation = .vertical
        stack.alignment = .leading
        stack.spacing = 6
        stack.edgeInsets = NSEdgeInsets(top: 8, left: 8, bottom: 8, right: 8)
        for v in [fieldBox, list] { v.widthAnchor.constraint(equalToConstant: width - 16).isActive = true }

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
        panel.setFrameOrigin(OptionsPanel.origin(for: size, near: caret))
        panel.orderFrontRegardless()
    }

    // MARK: Rows

    private func row(_ i: Int, _ r: QuranSuggestion, selected: Bool, lib: QuranLibrary,
                     mode: QuranInsert, translation: QuranTranslation?) -> NSView {
        var parts: [NSView] = []
        switch r {
        case .surah(let s):
            let label = NSTextField(labelWithString: "\(s.number)  \(s.name)  ·  \(s.english)  ·  \(s.ayahs) āyāt  ·  \(s.type)")
            label.font = .systemFont(ofSize: 13)
            let name = arabicLabel(s.arabic, lines: 1)
            let h = NSStackView(views: [label, NSView(), name])
            h.orientation = .horizontal
            parts.append(h)
        case .passage(let p), .hit(let p, _):
            let s = lib.surah(p.surah)
            let count = p.to - p.from + 1
            let ref = NSTextField(labelWithString: "\(p.reference)  \(s?.name ?? "")" + (count > 1 ? "  ·  \(count) āyāt" : ""))
            ref.font = .systemFont(ofSize: 11, weight: .medium)
            ref.textColor = .secondaryLabelColor
            parts.append(ref)
            parts.append(arabicLabel(lib.arabic(p.surah, p.from) ?? "", lines: selected ? 4 : 1))
            if mode != .arabic, selected, let t = translation, let tr = lib.translation(t.id, p.surah, p.from) {
                let l = NSTextField(wrappingLabelWithString: tr)
                l.font = .systemFont(ofSize: 12)
                l.textColor = .secondaryLabelColor
                l.maximumNumberOfLines = 3
                l.lineBreakMode = .byTruncatingTail
                l.preferredMaxLayoutWidth = width - 40
                parts.append(l)
            }
        case .problem(let text):
            let l = NSTextField(labelWithString: text)
            l.font = .systemFont(ofSize: 13)
            l.textColor = .secondaryLabelColor
            parts.append(l)
        }
        let v = NSStackView(views: parts)
        v.orientation = .vertical
        v.alignment = .leading
        v.spacing = 2
        for p in parts { p.widthAnchor.constraint(equalToConstant: width - 36).isActive = true }
        let click = ClickView { [weak self] in self?.onPick?(i) }
        click.wantsLayer = true
        click.layer?.cornerRadius = 6
        click.layer?.backgroundColor = selected ? NSColor.controlAccentColor.withAlphaComponent(0.22).cgColor : NSColor.clear.cgColor
        click.addSubview(v)
        v.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            v.leadingAnchor.constraint(equalTo: click.leadingAnchor, constant: 10),
            v.trailingAnchor.constraint(equalTo: click.trailingAnchor, constant: -10),
            v.topAnchor.constraint(equalTo: click.topAnchor, constant: 4),
            v.bottomAnchor.constraint(equalTo: click.bottomAnchor, constant: -4),
            click.widthAnchor.constraint(equalToConstant: width - 16),
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
        l.preferredMaxLayoutWidth = width - 40
        return l
    }

    // MARK: Pieces

    private func box(_ content: NSStackView, border: Bool) -> NSView {
        content.orientation = .horizontal
        content.spacing = 8
        let v = NSView()
        v.wantsLayer = true
        v.layer?.cornerRadius = 6
        v.layer?.borderWidth = 1
        v.layer?.borderColor = NSColor.controlAccentColor.withAlphaComponent(0.6).cgColor
        v.layer?.backgroundColor = NSColor.textBackgroundColor.withAlphaComponent(0.6).cgColor
        v.addSubview(content)
        content.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            content.leadingAnchor.constraint(equalTo: v.leadingAnchor, constant: 8),
            content.trailingAnchor.constraint(lessThanOrEqualTo: v.trailingAnchor, constant: -8),
            content.topAnchor.constraint(equalTo: v.topAnchor, constant: 5),
            content.bottomAnchor.constraint(equalTo: v.bottomAnchor, constant: -5),
        ])
        return v
    }

    private func small(_ s: String) -> NSTextField {
        let l = NSTextField(labelWithString: s)
        l.font = .systemFont(ofSize: 10)
        l.textColor = .tertiaryLabelColor
        return l
    }

    private func button(_ title: String, key: String?, action: @escaping () -> Void) -> NSView {
        let text = NSMutableAttributedString(string: title, attributes: [
            .font: NSFont.systemFont(ofSize: 11, weight: .medium), .foregroundColor: NSColor.labelColor])
        if let key {
            text.append(NSAttributedString(string: "  " + key, attributes: [
                .font: NSFont.systemFont(ofSize: 11), .foregroundColor: NSColor.secondaryLabelColor]))
        }
        let label = NSTextField(labelWithAttributedString: text)
        let v = ClickView(action: action)
        v.wantsLayer = true
        v.layer?.cornerRadius = 5
        v.layer?.borderWidth = 1
        v.layer?.borderColor = NSColor.separatorColor.cgColor
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

    private func separator() -> NSView {
        let b = NSBox()
        b.boxType = .separator
        return b
    }
}
