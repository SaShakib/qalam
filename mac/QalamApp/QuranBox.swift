import SwiftUI
import AppKit
import Carbon
import ApplicationServices
import QalamEngine

// The "Insert from the Qur'an" box (⌃⇧Q in any app, whatever keyboard is on).
// Qalam.app owns the shortcut (a Carbon hot key: no permission needed) and keeps running in the menu bar.
// The box is a normal window with a real search field, so the input source you have on decides what you
// type: English searches the translation and sūrah names, Qalam (Arabic) searches the Arabic.
// A result fills Sūrah / From / To; widen the range, then Insert. The text goes back into the app you
// were in: typed by the Qalam keyboard when it's on there, otherwise pasted (⌘V, needs Accessibility once).

// MARK: - The shortcut

final class QuranHotKey {
    static let shared = QuranHotKey()
    private var ref: EventHotKeyRef?
    private var handler: EventHandlerRef?
    private var current: Shortcut?

    /// (Re)registers the user's ⌃⇧Q (Settings → Shortcuts → Insert from the Qur'an).
    func register() {
        let s = SharedSettings.shortcut(for: .quran)
        guard s != current || (ref == nil && s != nil) else { return }
        if let ref { UnregisterEventHotKey(ref) }
        ref = nil
        current = s
        guard let s else { return }
        if handler == nil {
            var spec = EventTypeSpec(eventClass: OSType(kEventClassKeyboard), eventKind: UInt32(kEventHotKeyPressed))
            InstallEventHandler(GetApplicationEventTarget(), { _, _, _ in
                DispatchQueue.main.async { QuranBox.shared.toggle() }
                return noErr
            }, 1, &spec, nil, &handler)
        }
        var mods: UInt32 = 0
        if s.control { mods |= UInt32(controlKey) }
        if s.option { mods |= UInt32(optionKey) }
        if s.shift { mods |= UInt32(shiftKey) }
        if s.command { mods |= UInt32(cmdKey) }
        RegisterEventHotKey(UInt32(s.keyCode), mods, EventHotKeyID(signature: OSType(0x514C_4D51), id: 1),
                            GetApplicationEventTarget(), 0, &ref)
    }
}

// MARK: - The window

final class QuranBoxPanel: NSPanel {
    override var canBecomeKey: Bool { true }
    override func cancelOperation(_ sender: Any?) { QuranBox.shared.close(restore: true) }
}

final class QuranBox: NSObject, NSWindowDelegate {
    static let shared = QuranBox()

    private var panel: QuranBoxPanel?
    private var model: QuranBoxModel?
    /// The app you were typing in; it gets the focus (and the text) back.
    private var previous: NSRunningApplication?

    /// Shortcut, "open the box" from the keyboard, and the box requested at launch.
    func start() {
        QuranHotKey.shared.register()
        DistributedNotificationCenter.default().addObserver(forName: .init("com.qalam.quran.open"), object: nil, queue: .main) { _ in
            QuranBox.shared.open()
        }
    }

    func toggle() {
        if panel?.isVisible == true { close(restore: true) } else { open() }
    }

    func open() {
        guard let lib = QuranData.library else { NSSound.beep(); return }
        let front = NSWorkspace.shared.frontmostApplication
        if front?.bundleIdentifier != Bundle.main.bundleIdentifier { previous = front }
        let model = QuranBoxModel(lib: lib)
        self.model = model
        let view = QuranBoxView(model: model, insert: { [weak self] in self?.insert() },
                                cancel: { [weak self] in self?.close(restore: true) })
        let host = NSHostingView(rootView: view)
        let size = NSSize(width: 760, height: 620)
        let p = panel ?? QuranBoxPanel(contentRect: NSRect(origin: .zero, size: size),
                                       styleMask: [.titled, .closable, .resizable, .fullSizeContentView],
                                       backing: .buffered, defer: false)
        p.title = "Insert from the Qur'an"
        p.titlebarAppearsTransparent = true
        p.isFloatingPanel = true
        p.level = .floating
        p.hidesOnDeactivate = false
        p.isReleasedWhenClosed = false
        p.collectionBehavior = [.canJoinAllSpaces, .fullScreenAuxiliary]
        p.delegate = self
        p.contentView = host
        p.setContentSize(size)
        p.minSize = NSSize(width: 620, height: 460)
        // On the screen with the mouse, a little above the middle.
        let screen = NSScreen.screens.first { $0.frame.contains(NSEvent.mouseLocation) } ?? NSScreen.main
        if let f = screen?.visibleFrame {
            p.setFrameOrigin(NSPoint(x: f.midX - size.width / 2, y: f.midY - size.height / 2 + 60))
        }
        panel = p
        NSApp.activate(ignoringOtherApps: true)
        p.makeKeyAndOrderFront(nil)
    }

    /// `restore`: give the focus back to the app you were in (not when you clicked another app).
    func close(restore: Bool) {
        guard let p = panel, p.isVisible else { return }
        p.orderOut(nil)
        model = nil
        if restore { previous?.activate() }
    }

    func windowDidResignKey(_ notification: Notification) {
        // Clicked somewhere else: close, and leave the focus where you clicked.
        if NSApp.isActive { return }   // our own menus/pickers
        close(restore: false)
    }

    private func insert() {
        guard let model, let lib = QuranData.library else { return }
        let passage = model.passage
        let text = lib.text(passage, model.mode, translation: model.translation)
        SharedSettings.quranLast = passage.reference
        let target = previous
        close(restore: true)
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.25) { QuranInserter.insert(text, into: target) }
    }
}

// MARK: - Putting the text into the app you were in

enum QuranInserter {
    /// The Qalam keyboard types it when it's on in that app; otherwise ⌘V (Accessibility), or the clipboard.
    static func insert(_ text: String, into app: NSRunningApplication?) {
        let token = UUID().uuidString
        let center = DistributedNotificationCenter.default()
        var done = false
        let observer = center.addObserver(forName: .init("com.qalam.quran.inserted"), object: token, queue: .main) { _ in done = true }
        center.postNotificationName(.init("com.qalam.quran.insert"), object: nil,
                                    userInfo: ["text": text, "token": token, "bundle": app?.bundleIdentifier ?? ""],
                                    deliverImmediately: true)
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.35) {
            center.removeObserver(observer)
            if !done { paste(text) }
        }
    }

    static var canPaste: Bool { AXIsProcessTrusted() }

    /// Asks macOS to let Qalam paste for you (System Settings → Privacy & Security → Accessibility).
    static func askForAccessibility() {
        let key = kAXTrustedCheckOptionPrompt.takeUnretainedValue() as String
        _ = AXIsProcessTrustedWithOptions([key: true] as CFDictionary)
        NSWorkspace.shared.open(URL(string: "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility")!)
    }

    private static func paste(_ text: String) {
        let pb = NSPasteboard.general
        let saved = pb.pasteboardItems?.map { item -> [(NSPasteboard.PasteboardType, Data)] in
            item.types.compactMap { t in item.data(forType: t).map { (t, $0) } }
        } ?? []
        pb.clearContents()
        pb.setString(text, forType: .string)
        guard canPaste else {
            QuranToast.show("Copied. Press ⌘V to paste it.\nTo insert automatically, allow Qalam in Accessibility (Qalam → Settings).")
            return
        }
        let src = CGEventSource(stateID: .combinedSessionState)
        for down in [true, false] {
            let e = CGEvent(keyboardEventSource: src, virtualKey: 9, keyDown: down)   // V
            e?.flags = .maskCommand
            e?.post(tap: .cghidEventTap)
        }
        // Put back what was on the clipboard.
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
            guard !saved.isEmpty else { return }
            pb.clearContents()
            pb.writeObjects(saved.map { pairs -> NSPasteboardItem in
                let item = NSPasteboardItem()
                for (t, d) in pairs { item.setData(d, forType: t) }
                return item
            })
        }
    }
}

/// A short message near the top of the screen that fades by itself.
enum QuranToast {
    private static var panel: NSPanel?

    static func show(_ text: String) {
        let label = NSTextField(wrappingLabelWithString: text)
        label.font = .systemFont(ofSize: 13)
        label.alignment = .center
        let background = NSVisualEffectView()
        background.material = .hudWindow
        background.state = .active
        background.wantsLayer = true
        background.layer?.cornerRadius = 12
        background.addSubview(label)
        label.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            label.leadingAnchor.constraint(equalTo: background.leadingAnchor, constant: 18),
            label.trailingAnchor.constraint(equalTo: background.trailingAnchor, constant: -18),
            label.topAnchor.constraint(equalTo: background.topAnchor, constant: 12),
            label.bottomAnchor.constraint(equalTo: background.bottomAnchor, constant: -12),
            label.widthAnchor.constraint(equalToConstant: 420),
        ])
        let p = panel ?? NSPanel(contentRect: .zero, styleMask: [.nonactivatingPanel, .borderless], backing: .buffered, defer: true)
        p.level = .statusBar
        p.isOpaque = false
        p.backgroundColor = .clear
        p.hasShadow = true
        p.contentView = background
        let size = background.fittingSize
        p.setContentSize(size)
        if let f = (NSScreen.screens.first { $0.frame.contains(NSEvent.mouseLocation) } ?? NSScreen.main)?.visibleFrame {
            p.setFrameOrigin(NSPoint(x: f.midX - size.width / 2, y: f.maxY - size.height - 40))
        }
        p.alphaValue = 1
        p.orderFrontRegardless()
        panel = p
        DispatchQueue.main.asyncAfter(deadline: .now() + 4) {
            NSAnimationContext.runAnimationGroup({ $0.duration = 0.4; p.animator().alphaValue = 0 }) { p.orderOut(nil) }
        }
    }
}

// MARK: - What the box shows

final class QuranBoxModel: ObservableObject {
    let lib: QuranLibrary

    @Published var query = "" { didSet { if query != oldValue { search() } } }
    @Published private(set) var results: [QuranSuggestion] = []
    @Published private(set) var total = 0
    @Published private(set) var searchedIn = ""
    /// The translation a Latin or Bangla search looked in (its rows show that translation).
    private(set) var searchedTranslation: String?
    @Published var selected: Int? { didSet { applySelection() } }

    @Published var surah = 1 { didSet { if surah != oldValue && !applying { from = 1; to = 1 } } }
    @Published var from = 1 { didSet { if to < from { to = from } } }
    @Published var to = 1
    @Published var mode = SharedSettings.quranInsert { didSet { SharedSettings.quranInsert = mode } }
    @Published var translation = SharedSettings.quranTranslation { didSet { SharedSettings.quranTranslation = translation } }

    private var applying = false

    init(lib: QuranLibrary) {
        self.lib = lib
        if let last = SharedSettings.quranLast, let p = lib.resolve(last) {
            applying = true
            surah = p.surah; from = p.from; to = p.to
            applying = false
        }
    }

    var ayahCount: Int { lib.surah(surah)?.ayahs ?? 1 }
    var passage: Passage { Passage(surah: surah, from: from, to: max(from, to)) }

    // MARK: Search

    private static func script(_ s: String) -> String {
        if s.unicodeScalars.contains(where: { (0x0600...0x06FF).contains($0.value) || (0x0750...0x077F).contains($0.value) }) { return "arabic" }
        if s.unicodeScalars.contains(where: { (0x0980...0x09FF).contains($0.value) }) { return "bangla" }
        return "latin"
    }

    /// The translation to search for Latin or Bangla letters: the chosen one if it's that language.
    private func translationToSearch(_ script: String) -> String {
        let chosen = lib.translation(translation)
        if script == "bangla" { return chosen?.language == "Bangla" ? translation : "bn.bengali" }
        return chosen?.language == "English" ? translation : "en.sahih"
    }

    func search() {
        let q = query.trimmingCharacters(in: .whitespaces)
        guard !q.isEmpty else { results = []; total = 0; selected = nil; searchedIn = ""; return }
        var rows: [QuranSuggestion] = []
        // "2:255", "baqarah 255-257"
        if q.contains(where: \.isNumber), let p = lib.resolve(q), lib.parse(query: q).from != nil || Int(q) != nil {
            rows.append(.passage(p))
        }
        let script = Self.script(q)
        // Sūrah names (Latin or Arabic letters)
        if !q.contains(where: \.isNumber) {
            rows += lib.parse(query: q).surahs.prefix(4).map { .surah($0) }
        }
        var found = 0
        switch script {
        case "arabic":
            let r = lib.searchArabic(q, limit: 200)
            rows += r.hits.map { .hit($0, arabic: true) }
            found = r.total
            searchedIn = "the Arabic (harakat optional)"
            searchedTranslation = nil
        default:
            let id = translationToSearch(script)
            let r = lib.searchTranslation(id, q, limit: 200)
            rows += r.hits.map { .hit($0, arabic: false) }
            found = r.total
            searchedIn = lib.translation(id)?.title ?? id
            searchedTranslation = id
        }
        results = rows
        total = found
        selected = rows.isEmpty ? nil : 0
    }

    func move(_ d: Int) {
        guard !results.isEmpty else { return }
        selected = ((selected ?? -1) + d + results.count) % results.count
    }

    /// A result fills Sūrah / From / To.
    private func applySelection() {
        guard let i = selected, results.indices.contains(i) else { return }
        applying = true
        defer { applying = false }
        switch results[i] {
        case .surah(let s): surah = s.number; from = 1; to = 1
        case .passage(let p): surah = p.surah; from = p.from; to = p.to
        case .hit(let p, _): surah = p.surah; from = p.from; to = p.from
        case .problem: break
        }
    }

    func addBefore() { if from > 1 { from -= 1 } }
    func addAfter() { if to < ayahCount { to += 1 } }
}

// MARK: - The box's look

struct QuranBoxView: View {
    @ObservedObject var model: QuranBoxModel
    let insert: () -> Void
    let cancel: () -> Void
    @AppStorage("quranFont2") private var quranFontName = ArabicFonts.defaultQuran()
    @State private var canPaste = QuranInserter.canPaste

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            BoxSearchField(text: $model.query,
                           placeholder: "Search: a word in English or Arabic (switch to Qalam to type Arabic), a sūrah name, or 2:255",
                           onMove: { model.move($0) }, onReturn: insert, onEscape: {
                               if model.query.isEmpty { cancel() } else { model.query = "" }
                           })
                .frame(height: 30)

            if !model.query.trimmingCharacters(in: .whitespaces).isEmpty { resultsList }

            pickers
            preview
            Divider()
            bottom
        }
        .padding(.horizontal, 18)
        .padding(.bottom, 16)
        .padding(.top, 34)
        .frame(minWidth: 620, minHeight: 460)
        .onAppear { ArabicFonts.registerBundled() }
    }

    // MARK: Results

    private var resultsList: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(model.results.isEmpty ? "Nothing found." : "\(model.total) āyāt in \(model.searchedIn)"
                 + (model.total > 200 ? " (first 200)" : "") + ". ↑↓ or click to choose.")
                .font(.caption).foregroundStyle(.secondary)
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 2) {
                        ForEach(Array(model.results.enumerated()), id: \.offset) { i, r in
                            resultRow(r, selected: model.selected == i)
                                .id(i)
                                .contentShape(Rectangle())
                                .onTapGesture { model.selected = i }
                                .onTapGesture(count: 2) { model.selected = i; insert() }
                        }
                    }
                }
                .onChange(of: model.selected) { i in if let i { withAnimation(nil) { proxy.scrollTo(i) } } }
            }
            .frame(maxHeight: 200)
            .background(RoundedRectangle(cornerRadius: 8).fill(Color.secondary.opacity(0.06)))
        }
    }

    @ViewBuilder
    private func resultRow(_ r: QuranSuggestion, selected: Bool) -> some View {
        let lib = model.lib
        HStack(alignment: .firstTextBaseline, spacing: 10) {
            switch r {
            case .surah(let s):
                Text("Sūrah").font(.caption.weight(.semibold)).foregroundStyle(.secondary).frame(width: 70, alignment: .leading)
                Text("\(s.number). \(s.name)  ·  \(s.ayahs) āyāt")
                Spacer()
                Text(s.arabic).font(.custom(quranFontName, size: 18))
            case .passage(let p):
                Text(p.reference).font(.caption.weight(.semibold)).foregroundStyle(.secondary).frame(width: 70, alignment: .leading)
                Text("\(lib.surah(p.surah)?.name ?? "") \(p.reference)")
                Spacer()
            case .hit(let p, let arabic):
                Text(p.reference).font(.caption.weight(.semibold)).foregroundStyle(.secondary).frame(width: 70, alignment: .leading)
                if arabic {
                    Spacer(minLength: 0)
                    Text(shown(lib.arabic(p.surah, p.from) ?? "")).font(.custom(quranFontName, size: 19)).lineLimit(1)
                } else {
                    Text(lib.translation(model.translationShown(for: p), p.surah, p.from) ?? "").lineLimit(1)
                    Spacer(minLength: 0)
                }
            case .problem(let t):
                Text(t).foregroundStyle(.secondary)
            }
        }
        .padding(.horizontal, 10).padding(.vertical, 5)
        .background(RoundedRectangle(cornerRadius: 6).fill(selected ? Color.accentColor.opacity(0.25) : Color.clear))
    }

    // MARK: Sūrah / From / To

    private var pickers: some View {
        HStack(alignment: .bottom, spacing: 12) {
            labelled("Sūrah") {
                Picker("", selection: $model.surah) {
                    ForEach(model.lib.surahs) { s in Text("\(s.number). \(s.name)  ·  \(s.arabic)").tag(s.number) }
                }
                .labelsHidden()
                .frame(width: 280)
            }
            labelled("From āyah") {
                Picker("", selection: $model.from) {
                    ForEach(1...model.ayahCount, id: \.self) { Text("\($0)").tag($0) }
                }
                .labelsHidden().frame(width: 90)
            }
            labelled("To āyah") {
                Picker("", selection: $model.to) {
                    ForEach(model.from...model.ayahCount, id: \.self) { Text("\($0)").tag($0) }
                }
                .labelsHidden().frame(width: 90)
            }
            Button("+ āyah before") { model.addBefore() }.disabled(model.from <= 1)
            Button("+ āyah after") { model.addAfter() }.disabled(model.to >= model.ayahCount)
            Spacer(minLength: 0)
        }
    }

    private func labelled<V: View>(_ title: String, @ViewBuilder _ content: () -> V) -> some View {
        VStack(alignment: .leading, spacing: 3) {
            Text(title).font(.caption).foregroundStyle(.secondary)
            content()
        }
    }

    // MARK: Preview

    private var preview: some View {
        let p = model.passage
        let lib = model.lib
        let shownAyat = Array(p.from...min(p.to, p.from + 19))
        return VStack(alignment: .leading, spacing: 6) {
            Text("\(lib.surah(p.surah)?.name ?? "") \(p.reference)" + (p.to > p.from ? "  ·  \(p.to - p.from + 1) āyāt" : ""))
                .font(.caption.weight(.semibold)).foregroundStyle(.secondary)
            ScrollView {
                VStack(alignment: .leading, spacing: 10) {
                    ForEach(shownAyat, id: \.self) { a in
                        VStack(alignment: .leading, spacing: 4) {
                            if model.mode != .translation {
                                Text(shown(lib.arabic(p.surah, a) ?? "") + " " + number(a))
                                    .font(.custom(quranFontName, size: 24))
                                    .lineSpacing(8)
                                    .multilineTextAlignment(.trailing)
                                    .frame(maxWidth: .infinity, alignment: .trailing)
                            }
                            if model.mode != .arabic, let t = lib.translation(model.translation, p.surah, a) {
                                Text(t + " (\(p.surah):\(a))").font(.callout).foregroundStyle(.secondary)
                            }
                        }
                    }
                    if p.to > shownAyat.last! {
                        Text("… and \(p.to - shownAyat.last!) more āyāt").font(.caption).foregroundStyle(.secondary)
                    }
                }
                .padding(12)
            }
            .background(RoundedRectangle(cornerRadius: 8).fill(Color.secondary.opacity(0.06)))
        }
        .frame(maxHeight: .infinity)
    }

    // MARK: What to insert, and Insert

    private var bottom: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 12) {
                Picker("", selection: $model.mode) {
                    ForEach(QuranInsert.allCases) { Text($0.title).tag($0) }
                }
                .pickerStyle(.segmented).labelsHidden().frame(width: 380)
                if model.mode != .arabic {
                    Picker("", selection: $model.translation) {
                        ForEach(model.lib.available()) { Text($0.title).tag($0.id) }
                    }
                    .labelsHidden().frame(maxWidth: 300)
                }
                Spacer()
                Button("Cancel", action: cancel).keyboardShortcut(.cancelAction)
                Button("Insert", action: insert).keyboardShortcut(.defaultAction)
            }
            if !canPaste {
                HStack(spacing: 6) {
                    Text("With another keyboard on, Qalam pastes the text for you once it's allowed in Accessibility (until then it's copied, and you press ⌘V).")
                        .font(.caption).foregroundStyle(.secondary)
                    Button("Allow…") { QuranInserter.askForAccessibility() }.controlSize(.small)
                }
            }
        }
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didBecomeActiveNotification)) { _ in
            canPaste = QuranInserter.canPaste
        }
    }

    private func shown(_ s: String) -> String {
        quranFontName.uppercased().contains("KFGQPC") ? QuranLibrary.forKFGQPC(s) : s
    }

    /// The āyah-end mark: KFGQPC Hafs circles plain Arabic digits itself; other fonts need ۝ before them.
    private func number(_ n: Int) -> String {
        let digits = String(String(n).map { Character(Unicode.Scalar(0x0660 + Int(String($0))!)!) })
        return quranFontName.uppercased().contains("KFGQPC") ? digits : "\u{06DD}" + digits
    }
}

extension QuranBoxModel {
    /// For a translation hit, the translation it was found in.
    func translationShown(for p: Passage) -> String { searchedTranslation ?? translation }
}

// MARK: - The search field (↑↓ move through the results, Return inserts, Esc clears / closes)

struct BoxSearchField: NSViewRepresentable {
    @Binding var text: String
    let placeholder: String
    let onMove: (Int) -> Void
    let onReturn: () -> Void
    let onEscape: () -> Void

    func makeNSView(context: Context) -> NSSearchField {
        let f = NSSearchField()
        f.placeholderString = placeholder
        f.font = .systemFont(ofSize: 15)
        f.delegate = context.coordinator
        f.sendsSearchStringImmediately = true
        f.focusRingType = .default
        DispatchQueue.main.async { f.window?.makeFirstResponder(f) }
        return f
    }

    func updateNSView(_ f: NSSearchField, context: Context) {
        context.coordinator.parent = self
        if f.stringValue != text && (f.currentEditor() as? NSTextView)?.hasMarkedText() != true { f.stringValue = text }
    }

    func makeCoordinator() -> Coordinator { Coordinator(self) }

    final class Coordinator: NSObject, NSSearchFieldDelegate {
        var parent: BoxSearchField
        init(_ p: BoxSearchField) { parent = p }

        func controlTextDidChange(_ n: Notification) {
            guard let f = n.object as? NSSearchField else { return }
            parent.text = f.stringValue
        }

        func control(_ control: NSControl, textView: NSTextView, doCommandBy sel: Selector) -> Bool {
            if textView.hasMarkedText() { return false }   // the keyboard is still composing a word
            switch sel {
            case #selector(NSResponder.moveUp(_:)): parent.onMove(-1); return true
            case #selector(NSResponder.moveDown(_:)): parent.onMove(1); return true
            case #selector(NSResponder.insertNewline(_:)): parent.onReturn(); return true
            case #selector(NSResponder.cancelOperation(_:)): parent.onEscape(); return true
            default: return false
            }
        }
    }
}
