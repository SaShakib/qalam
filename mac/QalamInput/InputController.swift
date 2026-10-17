import Cocoa
import InputMethodKit
import QalamEngine

/// Receives every key while Qalam is the active input source.
/// The word being typed is shown underlined (marked text) and committed on space / punctuation.
@objc(QalamInputController)
final class QalamInputController: IMKInputController {
    private var composer = Composer()
    private let notFound = NSRange(location: NSNotFound, length: NSNotFound)

    override func activateServer(_ sender: Any!) {
        super.activateServer(sender)
        if !SharedSettings.keyboardWasEnabled { SharedSettings.keyboardWasEnabled = true }
    }

    override func recognizedEvents(_ sender: Any!) -> Int {
        Int(NSEvent.EventTypeMask.keyDown.rawValue)
    }

    override func handle(_ event: NSEvent!, client sender: Any!) -> Bool {
        guard let event = event, event.type == .keyDown, let client = sender as? IMKTextInput else { return false }
        let options = self.options(for: client)
        let flags = event.modifierFlags.intersection(.deviceIndependentFlagsMask)
        let busy = !composer.buffer.isEmpty

        // The user's shortcuts (Settings → Shortcuts): sukūn Off ↔ Smart (default ⌃⇧O), harakat Off ↔ On.
        let code = Int(event.keyCode)
        let hit: (Shortcut?) -> Bool = {
            $0?.matches(keyCode: code, control: flags.contains(.control), option: flags.contains(.option),
                        shift: flags.contains(.shift), command: flags.contains(.command)) ?? false
        }
        // The Qur'an box (⌃⇧Q) gets every key while it's open.
        if quranQuery != nil { return handleQuran(event, flags: flags, client: client, hit: hit) }

        if flags.contains(.control) || flags.contains(.option) {
            if hit(SharedSettings.shortcut(for: .quran)) {
                commitSelection(client)
                openQuran(client)
                return true
            }
            if hit(SharedSettings.shortcut(for: .sukun)) {
                toggleSukun(client)
                return true
            }
            if hit(SharedSettings.shortcut(for: .harakat)) {
                toggleHarakat(client)
                return true
            }
        }

        // Shortcuts (⌘C, ⌃A, ⌥…) finish the word and go to the app.
        if flags.contains(.command) || flags.contains(.control) || flags.contains(.option) {
            commitSelection(client)
            return false
        }

        // ↑ ↓ move through the options while a word is being typed.
        if busy && (event.keyCode == 125 || event.keyCode == 126) {
            move(event.keyCode == 125 ? 1 : -1, client)
            return true
        }

        let key: Composer.Key
        switch event.keyCode {
        case 51: key = .backspace
        case 53: key = .escape
        case 36, 76: key = .enter
        case 49: key = .space
        case 48, 115, 116, 117, 119, 121, 123, 124, 125, 126: key = .other   // tab, home, end, arrows…
        default:
            guard let raw = event.charactersIgnoringModifiers, raw.count == 1, var c = raw.first else {
                key = .other
                break
            }
            if c.isASCII && c.isLetter {
                // Case comes from Shift only, so Caps Lock can't turn س into ص by accident.
                c = Character(flags.contains(.shift) ? c.uppercased() : c.lowercased())
            } else if let chars = event.characters, chars.count == 1, let typed = chars.first {
                c = typed
            }
            key = .char(c)
        }

        // A key that ends the word puts in the highlighted option first.
        if busy && finishesWord(key) {
            commitSelection(client)
            if case .enter = key { return true }
        }

        let (consumed, actions) = composer.handle(key, options)
        for action in actions {
            switch action {
            case .mark: refresh(client, options)
            case .commit(let s): client.insertText(s, replacementRange: notFound)
            }
        }
        if composer.buffer.isEmpty {
            candidates = []
            OptionsPanel.shared.hide()
        }
        return consumed
    }

    private func finishesWord(_ key: Composer.Key) -> Bool {
        switch key {
        case .space, .enter, .other: return true
        case .char(let c): return !Qalam.accepts(c, buffer: composer.buffer)
        case .backspace, .escape: return false
        }
    }

    // MARK: Options

    private var candidates: [Candidate] = []
    private var selected = 0

    /// Recompute the options for the word being typed and show the highlighted one in place.
    private func refresh(_ client: IMKTextInput, _ o: Options) {
        let typed = composer.buffer
        guard !typed.isEmpty else {
            setMarked("", client)
            candidates = []
            OptionsPanel.shared.hide()
            return
        }
        candidates = Qalam.candidates(typed, o)
        selected = defaultSelection(typed)
        setMarked(candidates[selected].text, client)
        showPanel(client, o)
    }

    /// What the user picked for this word before; otherwise their harakat habit; otherwise as typed.
    private func defaultSelection(_ typed: String) -> Int {
        if let learned = SharedSettings.learnedChoice(for: typed),
           let i = candidates.firstIndex(where: { Self.plainZWJ($0.text) == learned }) {
            return i
        }
        if SharedSettings.prefersPlain, let i = candidates.firstIndex(where: { $0.kind == .plain }) {
            return i
        }
        return 0
    }

    private static func plainZWJ(_ s: String) -> String { s.replacingOccurrences(of: "\u{200D}", with: "") }

    private func move(_ step: Int, _ client: IMKTextInput) {
        guard !candidates.isEmpty else { return }
        selected = (selected + step + candidates.count) % candidates.count
        setMarked(candidates[selected].text, client)
        showPanel(client, options(for: client))
    }

    /// Insert the highlighted option for the word being typed, and remember the choice.
    private func commitSelection(_ client: IMKTextInput) {
        let typed = composer.buffer
        guard !typed.isEmpty else { return }
        if candidates.isEmpty {
            candidates = Qalam.candidates(typed, options(for: client))
            selected = 0
        }
        let chosen = candidates[min(selected, candidates.count - 1)]
        client.insertText(chosen.text, replacementRange: notFound)
        SharedSettings.remember(Candidate(text: Self.plainZWJ(chosen.text), kind: chosen.kind, label: chosen.label),
                                for: typed, among: candidates)
        composer.reset()
        candidates = []
        selected = 0
        OptionsPanel.shared.hide()
    }

    private func setMarked(_ s: String, _ client: IMKTextInput) {
        let text = NSAttributedString(string: s, attributes: [.underlineStyle: NSUnderlineStyle.single.rawValue])
        client.setMarkedText(text, selectionRange: NSRange(location: (s as NSString).length, length: 0),
                             replacementRange: notFound)
    }

    private func showPanel(_ client: IMKTextInput, _ o: Options) {
        guard SharedSettings.showOptions, !candidates.isEmpty else { OptionsPanel.shared.hide(); return }
        var caret = NSRect.zero
        _ = client.attributes(forCharacterIndex: 0, lineHeightRectangle: &caret)
        let panel = OptionsPanel.shared
        panel.onPick = { [weak self] i in
            guard let self, let c = self.client() else { return }
            self.selected = i
            self.commitSelection(c)
        }
        panel.onToggleSukun = { [weak self] in
            guard let self, let c = self.client() else { return }
            self.toggleSukun(c)
        }
        panel.onToggleHarakat = { [weak self] in
            guard let self, let c = self.client() else { return }
            self.toggleHarakat(c)
        }
        panel.show(candidates, selected: selected, sukunOff: o.sukun == .off, plain: o.harakat == .none, near: caret,
                   sukunKey: SharedSettings.shortcut(for: .sukun)?.display,
                   harakatKey: SharedSettings.shortcut(for: .harakat)?.display)
    }

    /// Sukūn Off ↔ Smart (button in the panel, or the sukūn shortcut).
    private func toggleSukun(_ client: IMKTextInput) {
        change { $0.sukun = $0.sukun == .off ? .smart : .off }
        if !composer.buffer.isEmpty { refresh(client, options(for: client)) }
    }

    /// Harakat Off ↔ Full (button in the panel, or the harakat shortcut).
    private func toggleHarakat(_ client: IMKTextInput) {
        change { $0.harakat = $0.harakat == .none ? .full : .none }
        if !composer.buffer.isEmpty { refresh(client, options(for: client)) }
    }

    /// Settings for this keystroke. Chrome-based apps get the Allah-ligature guard (see Options).
    private func options(for client: IMKTextInput) -> Options {
        var o = SharedSettings.load()
        o.blockAllahLigature = Self.isChromiumBased(client.bundleIdentifier())
        return o
    }

    private static var chromiumCache: [String: Bool] = [:]

    /// Chrome, Edge, Brave, Arc and Electron apps (Claude, VS Code, Slack…) all ship Chromium's
    /// .pak resource files; native apps (Safari, Notes, Pages…) don't.
    static func isChromiumBased(_ bundleID: String?) -> Bool {
        guard let id = bundleID, !id.isEmpty else { return false }
        if let known = chromiumCache[id] { return known }
        var found = false
        if let app = NSWorkspace.shared.urlForApplication(withBundleIdentifier: id) {
            let fm = FileManager.default
            let frameworks = app.appendingPathComponent("Contents/Frameworks")
            let names = (try? fm.contentsOfDirectory(atPath: frameworks.path)) ?? []
            for fw in names where fw.hasSuffix(".framework") {
                let res = frameworks.appendingPathComponent(fw).appendingPathComponent("Resources").path
                if let files = try? fm.contentsOfDirectory(atPath: res), files.contains(where: { $0.hasSuffix(".pak") }) {
                    found = true
                    break
                }
            }
        }
        chromiumCache[id] = found
        return found
    }

    override func commitComposition(_ sender: Any!) {
        guard let client = sender as? IMKTextInput else { return }
        commitSelection(client)
    }

    override func deactivateServer(_ sender: Any!) {
        closeQuran()
        commitComposition(sender)
        super.deactivateServer(sender)
    }

    // MARK: The Qur'an box

    /// What's typed in the box; nil while it's closed.
    private var quranQuery: String?
    private var quranRows: [QuranSuggestion] = []
    private var quranSelected = 0
    private var quranTotal: Int?

    private func openQuran(_ client: IMKTextInput) {
        guard QuranStore.shared != nil else { NSSound.beep(); return }
        quranQuery = ""
        quranRows = []
        quranSelected = 0
        refreshQuran(client)
    }

    private func closeQuran() {
        quranQuery = nil
        quranRows = []
        QuranPanel.shared.hide()
    }

    private func handleQuran(_ event: NSEvent, flags: NSEvent.ModifierFlags, client: IMKTextInput,
                             hit: (Shortcut?) -> Bool) -> Bool {
        guard var q = quranQuery else { return false }
        if hit(SharedSettings.shortcut(for: .quran)) { closeQuran(); return true }     // the shortcut again closes it
        if flags.contains(.command) { closeQuran(); return false }                       // ⌘ shortcuts go to the app
        switch event.keyCode {
        case 53: closeQuran(); return true                                               // Esc
        case 36, 76: pickQuran(quranSelected, client); return true                       // Return
        case 125, 126:                                                                   // ↓ ↑
            if !quranRows.isEmpty {
                quranSelected = (quranSelected + (event.keyCode == 125 ? 1 : -1) + quranRows.count) % quranRows.count
                showQuran(client)
            }
            return true
        case 48:                                                                         // Tab, ⇧Tab
            flags.contains(.shift) ? cycleTranslation() : cycleQuranMode()
            showQuran(client)
            return true
        case 51:                                                                         // ⌫
            if q.isEmpty { closeQuran(); return true }
            q.removeLast()
        default:
            guard let typed = event.characters, !typed.isEmpty,
                  typed.unicodeScalars.allSatisfy({ !CharacterSet.controlCharacters.contains($0) && $0.value < 0xF700 })
            else { return true }
            q += typed
        }
        quranQuery = q
        refreshQuran(client)
        return true
    }

    /// New dropdown for the typed text. Latin that isn't a reference is also read as Arabic (rahmaan → رحمان).
    private func refreshQuran(_ client: IMKTextInput) {
        guard let lib = QuranStore.shared, let q = quranQuery else { return }
        var plain = Options()
        plain.harakat = .none
        let phonetic = q.contains(where: { $0.isLetter && $0.isASCII }) ? Qalam.text(q, plain) : nil
        quranRows = lib.suggestions(for: q, arabic: phonetic, limit: 6)
        quranTotal = nil
        if quranRows.contains(where: { if case .hit = $0 { return true }; return false }) {
            let ar = q.unicodeScalars.contains { (0x0600...0x06FF).contains($0.value) } ? q : (phonetic ?? q)
            quranTotal = lib.searchArabic(ar, phonetic: phonetic != nil, limit: 0).total
        }
        quranSelected = 0
        showQuran(client)
    }

    private func showQuran(_ client: IMKTextInput) {
        guard let lib = QuranStore.shared, let q = quranQuery else { return }
        var caret = NSRect.zero
        _ = client.attributes(forCharacterIndex: 0, lineHeightRectangle: &caret)
        let panel = QuranPanel.shared
        panel.onPick = { [weak self] i in
            guard let self, let c = self.client() else { return }
            self.pickQuran(i, c)
        }
        panel.onCycleMode = { [weak self] in
            guard let self, let c = self.client() else { return }
            self.cycleQuranMode(); self.showQuran(c)
        }
        panel.onCycleTranslation = { [weak self] in
            guard let self, let c = self.client() else { return }
            self.cycleTranslation(); self.showQuran(c)
        }
        panel.show(query: q, rows: quranRows, selected: quranSelected, total: quranTotal, library: lib,
                   mode: SharedSettings.quranInsert, translation: lib.translation(SharedSettings.quranTranslation),
                   near: caret, key: SharedSettings.shortcut(for: .quran)?.display)
    }

    /// A sūrah row fills in "2:" for the āyah number; a passage or search result is inserted.
    private func pickQuran(_ i: Int, _ client: IMKTextInput) {
        guard let lib = QuranStore.shared, i < quranRows.count else {
            if let lib = QuranStore.shared, let q = quranQuery, let p = lib.resolve(q) { insertQuran(p, lib, client) }
            return
        }
        switch quranRows[i] {
        case .surah(let s):
            quranQuery = "\(s.number):"
            refreshQuran(client)
        case .passage(let p), .hit(let p, _):
            insertQuran(p, lib, client)
        case .problem:
            NSSound.beep()
        }
    }

    private func insertQuran(_ p: Passage, _ lib: QuranLibrary, _ client: IMKTextInput) {
        let text = lib.text(p, SharedSettings.quranInsert, translation: SharedSettings.quranTranslation)
        closeQuran()
        client.insertText(text, replacementRange: notFound)
    }

    private func cycleQuranMode() {
        let all = QuranInsert.allCases
        SharedSettings.quranInsert = all[(all.firstIndex(of: SharedSettings.quranInsert)! + 1) % all.count]
    }

    /// The next translation on this Mac (bundled ones and those downloaded in the app).
    private func cycleTranslation() {
        guard let have = QuranStore.shared?.available().map(\.id), !have.isEmpty else { return }
        let i = have.firstIndex(of: SharedSettings.quranTranslation) ?? -1
        SharedSettings.quranTranslation = have[(i + 1) % have.count]
        if SharedSettings.quranInsert == .arabic { SharedSettings.quranInsert = .both }
    }

    // MARK: Input menu (the ق icon in the menu bar)

    override func menu() -> NSMenu! {
        let o = SharedSettings.load()
        let m = NSMenu(title: "Qalam")
        m.addItem(item("Open Qalam (letters, words, practice)…", #selector(openApp(_:))))
        m.addItem(item("Insert from the Qur'an…" + Self.keyHint(.quran), #selector(quranFromMenu(_:))))
        m.addItem(item("Qur'an page (read, search, translations)…", #selector(openQuranPage(_:))))
        m.addItem(.separator())
        m.addItem(item("Everyday style", #selector(styleEveryday(_:)), on: o.style == .everyday))
        m.addItem(item("Qur'an style (ٱ ـٰ ـٓ)", #selector(styleQuran(_:)), on: o.style == .quran))
        m.addItem(.separator())
        m.addItem(item("Full harakat", #selector(harakatFull(_:)), on: o.harakat == .full))
        m.addItem(item("Only the harakat I type", #selector(harakatTyped(_:)), on: o.harakat == .asTyped))
        m.addItem(item("No harakat" + Self.keyHint(.harakat), #selector(harakatNone(_:)), on: o.harakat == .none))
        m.addItem(.separator())
        m.addItem(item("Sukūn: Smart (only where needed)", #selector(sukunSmart(_:)), on: o.sukun == .smart))
        m.addItem(item("Sukūn: Full (every stop)", #selector(sukunFull(_:)), on: o.sukun == .full))
        m.addItem(item("Sukūn: Off" + Self.keyHint(.sukun), #selector(sukunOff(_:)), on: o.sukun == .off))
        m.addItem(item("Change shortcuts…", #selector(openShortcuts(_:))))
        m.addItem(.separator())
        m.addItem(item("Show options while typing", #selector(toggleOptions(_:)), on: SharedSettings.showOptions))
        m.addItem(item("Arabic digits ١٢٣", #selector(toggleDigits(_:)), on: o.arabicDigits))
        m.addItem(item("Smart spelling (اللَّه، هَٰذَا)", #selector(toggleSpelling(_:)), on: o.spellingWords))
        return m
    }

    private static func keyHint(_ a: Shortcut.Action) -> String {
        SharedSettings.shortcut(for: a).map { "   " + $0.display } ?? ""
    }

    private func item(_ title: String, _ action: Selector, on: Bool = false) -> NSMenuItem {
        let i = NSMenuItem(title: title, action: action, keyEquivalent: "")
        i.target = self
        i.state = on ? .on : .off
        return i
    }

    private func change(_ f: (inout Options) -> Void) {
        var o = SharedSettings.load()
        f(&o)
        SharedSettings.save(o)
    }

    @objc func styleEveryday(_ sender: Any?) { change { $0.style = .everyday } }
    @objc func styleQuran(_ sender: Any?) { change { $0.style = .quran } }
    @objc func harakatFull(_ sender: Any?) { change { $0.harakat = .full } }
    @objc func harakatTyped(_ sender: Any?) { change { $0.harakat = .asTyped } }
    @objc func harakatNone(_ sender: Any?) { change { $0.harakat = .none } }
    @objc func toggleDigits(_ sender: Any?) { change { $0.arabicDigits.toggle() } }
    @objc func sukunSmart(_ sender: Any?) { change { $0.sukun = .smart } }
    @objc func sukunFull(_ sender: Any?) { change { $0.sukun = .full } }
    @objc func sukunOff(_ sender: Any?) { change { $0.sukun = .off } }
    @objc func toggleOptions(_ sender: Any?) { SharedSettings.showOptions.toggle() }
    @objc func toggleSpelling(_ sender: Any?) { change { $0.spellingWords.toggle() } }

    @objc func quranFromMenu(_ sender: Any?) {
        if let c = client() { openQuran(c) }
    }

    @objc func openQuranPage(_ sender: Any?) {
        SharedSettings.requestPage("quran")
        openApp(sender)
    }

    @objc func openShortcuts(_ sender: Any?) {
        SharedSettings.requestPage("settings")
        openApp(sender)
    }

    @objc func openApp(_ sender: Any?) {
        let ws = NSWorkspace.shared
        let url = ws.urlForApplication(withBundleIdentifier: "com.qalam.app")
            ?? FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Applications/Qalam.app")
        ws.openApplication(at: url, configuration: NSWorkspace.OpenConfiguration(), completionHandler: nil)
    }
}
