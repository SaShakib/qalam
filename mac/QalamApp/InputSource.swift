import AppKit
import Carbon

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

    static var isEnabled: Bool { !sources(includeDisabled: false).isEmpty }

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
