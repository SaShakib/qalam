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
        let options = SharedSettings.load()
        let flags = event.modifierFlags.intersection(.deviceIndependentFlagsMask)

        // Shortcuts (⌘C, ⌃A, ⌥…) finish the word and go to the app.
        if flags.contains(.command) || flags.contains(.control) || flags.contains(.option) {
            perform(composer.flushActions(options), client)
            return false
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

        let (consumed, actions) = composer.handle(key, options)
        perform(actions, client)
        return consumed
    }

    private func perform(_ actions: [Composer.Action], _ client: IMKTextInput) {
        for action in actions {
            switch action {
            case .mark(let s):
                let text = NSAttributedString(string: s, attributes: [
                    .underlineStyle: NSUnderlineStyle.single.rawValue,
                ])
                client.setMarkedText(text, selectionRange: NSRange(location: (s as NSString).length, length: 0),
                                     replacementRange: notFound)
            case .commit(let s):
                client.insertText(s, replacementRange: notFound)
            }
        }
    }

    override func commitComposition(_ sender: Any!) {
        guard let client = sender as? IMKTextInput else { return }
        perform(composer.flushActions(SharedSettings.load()), client)
    }

    override func deactivateServer(_ sender: Any!) {
        commitComposition(sender)
        super.deactivateServer(sender)
    }

    // MARK: Input menu (the ق icon in the menu bar)

    override func menu() -> NSMenu! {
        let o = SharedSettings.load()
        let m = NSMenu(title: "Qalam")
        m.addItem(item("Open Qalam (letters, words, practice)…", #selector(openApp(_:))))
        m.addItem(.separator())
        m.addItem(item("Everyday style", #selector(styleEveryday(_:)), on: o.style == .everyday))
        m.addItem(item("Qur'an style (ٱ ـٰ ـٓ)", #selector(styleQuran(_:)), on: o.style == .quran))
        m.addItem(.separator())
        m.addItem(item("Full harakat (automatic sukūn)", #selector(harakatFull(_:)), on: o.harakat == .full))
        m.addItem(item("Only the harakat I type", #selector(harakatTyped(_:)), on: o.harakat == .asTyped))
        m.addItem(item("No harakat", #selector(harakatNone(_:)), on: o.harakat == .none))
        m.addItem(.separator())
        m.addItem(item("Arabic digits ١٢٣", #selector(toggleDigits(_:)), on: o.arabicDigits))
        m.addItem(item("Smart spelling (اللَّه، هَٰذَا)", #selector(toggleSpelling(_:)), on: o.spellingWords))
        return m
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
    @objc func toggleSpelling(_ sender: Any?) { change { $0.spellingWords.toggle() } }

    @objc func openApp(_ sender: Any?) {
        let ws = NSWorkspace.shared
        let url = ws.urlForApplication(withBundleIdentifier: "com.qalam.app")
            ?? FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Applications/Qalam.app")
        ws.openApplication(at: url, configuration: NSWorkspace.OpenConfiguration(), completionHandler: nil)
    }
}
