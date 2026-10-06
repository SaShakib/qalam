import AppKit
import Carbon
import QalamEngine

/// Finds the Qalam keyboard among the Mac's input sources.
enum InputSource {
    static let bundleID = "com.qalam.inputmethod.Qalam"

    /// The keyboard is in ~/Library/Input Methods (Homebrew, install script) or /Library/Input Methods (.pkg).
    static var bundleURL: URL {
        let user = FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Library/Input Methods/QalamInput.app")
        let system = URL(fileURLWithPath: "/Library/Input Methods/QalamInput.app")
        return FileManager.default.fileExists(atPath: user.path) ? user : system
    }

    static var isInstalled: Bool { FileManager.default.fileExists(atPath: bundleURL.path) }

    static func sources(includeDisabled: Bool) -> [TISInputSource] {
        let filter = [kTISPropertyBundleID as String: bundleID] as CFDictionary
        guard let list = TISCreateInputSourceList(filter, includeDisabled)?.takeRetainedValue() else { return [] }
        return list as! [TISInputSource]
    }

    static var isEnabled: Bool {
        let on = !sources(includeDisabled: false).isEmpty
        if on && !SharedSettings.keyboardWasEnabled { SharedSettings.keyboardWasEnabled = true }
        return on
    }

    static var rememberedEnabled: Bool { SharedSettings.keyboardWasEnabled }

    /// macOS drops an input source from the menu when its files are replaced (reinstall, update,
    /// brew upgrade). If the user had turned Qalam on before, turn it back on.
    static func restoreIfNeeded() {
        guard rememberedEnabled, isInstalled, !isEnabled else { return }
        enable()
    }

    /// Adds Qalam to the input menu (same as System Settings → Input Sources → +).
    @discardableResult
    static func enable() -> Bool {
        TISRegisterInputSource(bundleURL as CFURL)
        for s in sources(includeDisabled: true) { TISEnableInputSource(s) }
        return isEnabled
    }

    /// Switches the current keyboard to Qalam.
    static func select() {
        for s in sources(includeDisabled: false) { TISSelectInputSource(s) }
    }

    static func openKeyboardSettings() {
        if let url = URL(string: "x-apple.systempreferences:com.apple.Keyboard-Settings.extension") {
            NSWorkspace.shared.open(url)
        }
    }
}
