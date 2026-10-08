import Cocoa
import CoreText
import QalamEngine

/// The options panel shown under the word while typing (like Avro's suggestion list).
/// A non-activating panel: it never takes focus from the app you are typing in, but rows
/// and buttons can still be clicked.
final class OptionsPanel {
    static let shared = OptionsPanel()

    var onPick: ((Int) -> Void)?
    var onToggleSukun: (() -> Void)?
    var onToggleHarakat: (() -> Void)?

    private var panel: NSPanel?
    private lazy var arabicFont: NSFont = {
        Self.registerQalamFonts()
        return NSFont(name: "NotoNaskhArabic-Regular", size: 22) ?? NSFont.systemFont(ofSize: 22)
    }()

    func hide() { panel?.orderOut(nil) }

    /// For layout checks only.
    var snapshotView: NSView? { panel?.contentView }

    func show(_ candidates: [Candidate], selected: Int, sukunOff: Bool, plain: Bool, near caret: NSRect,
              sukunKey: String?, harakatKey: String?) {
        let panel = self.panel ?? makePanel()
        self.panel = panel

        let rows = NSStackView()
        rows.orientation = .vertical
        rows.alignment = .leading
        rows.spacing = 2
        for (i, c) in candidates.enumerated() {
            rows.addArrangedSubview(row(i, c, selected: i == selected))
        }

        let buttons = NSStackView(views: [
            button(sukunOff ? "Sukūn: Off" : "Sukūn: Smart", key: sukunKey) { [weak self] in self?.onToggleSukun?() },
            button(plain ? "Harakat: Off" : "Harakat: On", key: harakatKey) { [weak self] in self?.onToggleHarakat?() },
        ])
        buttons.orientation = .horizontal
        buttons.spacing = 6

        let hint = NSTextField(labelWithString: "↑↓ choose · Space to insert · click to pick")
        hint.font = .systemFont(ofSize: 10)
        hint.textColor = .tertiaryLabelColor

        let stack = NSStackView(views: [rows, separator(), buttons, hint])
        stack.orientation = .vertical
        stack.alignment = .leading
        stack.spacing = 6
        stack.edgeInsets = NSEdgeInsets(top: 8, left: 8, bottom: 8, right: 8)

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

        let size = stack.fittingSize
        let width = max(size.width, 220)
        panel.contentView = background
        panel.setContentSize(NSSize(width: width, height: size.height))
        panel.setFrameOrigin(origin(for: NSSize(width: width, height: size.height), near: caret))
        panel.orderFrontRegardless()
    }

    // MARK: Layout

    private func makePanel() -> NSPanel {
        let p = NSPanel(contentRect: NSRect(x: 0, y: 0, width: 240, height: 120),
                        styleMask: [.nonactivatingPanel, .borderless], backing: .buffered, defer: true)
        p.level = .popUpMenu
        p.isFloatingPanel = true
        p.hidesOnDeactivate = false
        p.becomesKeyOnlyIfNeeded = true
        p.isOpaque = false
        p.backgroundColor = .clear
        p.hasShadow = true
        p.collectionBehavior = [.canJoinAllSpaces, .fullScreenAuxiliary, .transient]
        return p
    }

    /// Just under the word; above it if there's no room below. Near the mouse if the app gives no caret.
    private func origin(for size: NSSize, near caret: NSRect) -> NSPoint {
        var anchor = caret
        if anchor.origin == .zero && anchor.size == .zero {
            let m = NSEvent.mouseLocation
            anchor = NSRect(x: m.x, y: m.y - 10, width: 1, height: 18)
        }
        let screen = NSScreen.screens.first { $0.frame.contains(anchor.origin) } ?? NSScreen.main
        let visible = screen?.visibleFrame ?? NSRect(x: 0, y: 0, width: 1440, height: 900)
        var x = anchor.minX
        var y = anchor.minY - size.height - 6
        if y < visible.minY { y = anchor.maxY + 6 }
        x = min(max(x, visible.minX + 4), visible.maxX - size.width - 4)
        y = min(max(y, visible.minY + 4), visible.maxY - size.height - 4)
        return NSPoint(x: x, y: y)
    }

    private func row(_ i: Int, _ c: Candidate, selected: Bool) -> NSView {
        let number = NSTextField(labelWithString: "\(i + 1)")
        number.font = .monospacedDigitSystemFont(ofSize: 12, weight: .medium)
        number.textColor = .secondaryLabelColor
        let text = NSTextField(labelWithString: c.text)
        text.font = arabicFont
        let label = NSTextField(labelWithString: c.label)
        label.font = .systemFont(ofSize: 11)
        label.textColor = .secondaryLabelColor
        let h = NSStackView(views: [number, text, label])
        h.orientation = .horizontal
        h.spacing = 10
        h.edgeInsets = NSEdgeInsets(top: 2, left: 8, bottom: 2, right: 10)

        let v = ClickView { [weak self] in self?.onPick?(i) }
        v.wantsLayer = true
        v.layer?.cornerRadius = 6
        v.layer?.backgroundColor = selected ? NSColor.controlAccentColor.withAlphaComponent(0.22).cgColor : NSColor.clear.cgColor
        v.addSubview(h)
        h.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            h.leadingAnchor.constraint(equalTo: v.leadingAnchor),
            h.trailingAnchor.constraint(equalTo: v.trailingAnchor),
            h.topAnchor.constraint(equalTo: v.topAnchor),
            h.bottomAnchor.constraint(equalTo: v.bottomAnchor),
        ])
        return v
    }

    /// A button with its shortcut shown after the title (e.g. "Sukūn: Smart  ⌃⇧O").
    private func button(_ title: String, key: String?, action: @escaping () -> Void) -> NSView {
        let text = NSMutableAttributedString(string: title, attributes: [
            .font: NSFont.systemFont(ofSize: 11, weight: .medium), .foregroundColor: NSColor.labelColor])
        if let key {
            text.append(NSAttributedString(string: "  " + key, attributes: [
                .font: NSFont.systemFont(ofSize: 11), .foregroundColor: NSColor.secondaryLabelColor]))
        }
        let label = NSTextField(labelWithAttributedString: text)
        let v = ClickView(action: action)
        v.toolTip = key.map { "Shortcut: \($0). Change it in Qalam → Settings → Shortcuts." }
            ?? "Set a shortcut in Qalam → Settings → Shortcuts."
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

    /// Use the fonts that ship inside Qalam.app (Noto Naskh Arabic, Amiri Quran…).
    static func registerQalamFonts() {
        guard let app = NSWorkspace.shared.urlForApplication(withBundleIdentifier: "com.qalam.app") else { return }
        let dir = app.appendingPathComponent("Contents/Resources/Fonts")
        let files = (try? FileManager.default.contentsOfDirectory(at: dir, includingPropertiesForKeys: nil)) ?? []
        for f in files where f.pathExtension == "ttf" {
            CTFontManagerRegisterFontsForURL(f as CFURL, .process, nil)
        }
    }
}

/// A view that reacts to a click even though its window is never the key window.
final class ClickView: NSView {
    private let action: () -> Void
    init(action: @escaping () -> Void) {
        self.action = action
        super.init(frame: .zero)
    }
    required init?(coder: NSCoder) { fatalError() }
    override func acceptsFirstMouse(for event: NSEvent?) -> Bool { true }
    override func mouseDown(with event: NSEvent) { action() }
}
