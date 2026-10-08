import Foundation

/// A keyboard shortcut the user can change (Settings → Shortcuts), stored in the shared settings
/// so the keyboard and the app agree. `keyCode` is the Mac virtual key code.
public struct Shortcut: Equatable, Codable {
    public var keyCode: Int
    public var key: String          // what to show for the key, e.g. "O"
    public var control = false
    public var option = false
    public var shift = false
    public var command = false

    public init(keyCode: Int, key: String, control: Bool = false, option: Bool = false, shift: Bool = false, command: Bool = false) {
        self.keyCode = keyCode
        self.key = key
        self.control = control
        self.option = option
        self.shift = shift
        self.command = command
    }

    /// Mac symbols, in Apple's order: ⌃⌥⇧⌘.
    public var display: String {
        (control ? "⌃" : "") + (option ? "⌥" : "") + (shift ? "⇧" : "") + (command ? "⌘" : "") + key
    }

    /// A shortcut must use Control or Option, so it never swallows a letter you are typing.
    /// (⌘ shortcuts are taken by the app's menus before a keyboard ever sees them.)
    public var isUsable: Bool { (control || option) && !command }

    public func matches(keyCode: Int, control: Bool, option: Bool, shift: Bool, command: Bool) -> Bool {
        keyCode == self.keyCode && control == self.control && option == self.option
            && shift == self.shift && command == self.command
    }

    public enum Action: String, CaseIterable, Identifiable {
        case sukun, harakat
        public var id: String { rawValue }
        public var title: String {
            switch self {
            case .sukun: return "Sukūn off / on"
            case .harakat: return "Harakat off / on"
            }
        }
        /// Control-Shift-O for sukūn; harakat has none until the user sets one.
        public var defaultShortcut: Shortcut? {
            switch self {
            case .sukun: return Shortcut(keyCode: 31, key: "O", control: true, shift: true)
            case .harakat: return nil
            }
        }
    }
}

extension SharedSettings {
    /// The shortcut for an action, or nil when the user removed it.
    public static func shortcut(for action: Shortcut.Action) -> Shortcut? {
        let key = "shortcut." + action.rawValue
        guard let data = defaults.data(forKey: key) else { return action.defaultShortcut }
        return try? JSONDecoder().decode(Shortcut.self, from: data)   // "null" → removed
    }

    public static func setShortcut(_ s: Shortcut?, for action: Shortcut.Action) {
        let key = "shortcut." + action.rawValue
        if let data = try? JSONEncoder().encode(s) { defaults.set(data, forKey: key) }
        defaults.synchronize()
    }

    public static func resetShortcut(for action: Shortcut.Action) {
        defaults.removeObject(forKey: "shortcut." + action.rawValue)
        defaults.synchronize()
    }

    /// A page the keyboard asked the app to open (e.g. "settings" from the menu's "Change shortcuts…").
    /// The app reads it once when it becomes active.
    public static func takeRequestedPage() -> String? {
        let p = defaults.string(forKey: "openPage")
        if p != nil { defaults.removeObject(forKey: "openPage") }
        return p
    }

    public static func requestPage(_ page: String) {
        defaults.set(page, forKey: "openPage")
        defaults.synchronize()
    }

    /// The other action already using this shortcut, if any.
    public static func action(using s: Shortcut, except: Shortcut.Action) -> Shortcut.Action? {
        Shortcut.Action.allCases.first { $0 != except && shortcut(for: $0) == s }
    }
}
