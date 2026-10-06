import Foundation

/// Turns key presses into "show this underlined" / "commit this" actions.
/// The macOS input method and the app's converter both run on this.
public struct Composer {
    public enum Key {
        case char(Character)
        case backspace, escape, enter, space
        case other        // tab, arrows, … : finish the word and let the key through
    }

    public enum Action: Equatable {
        case mark(String)     // replace the underlined preview
        case commit(String)   // insert final text
    }

    public private(set) var buffer = ""

    public init() {}

    /// Returns whether the key was used up, plus what to do.
    public mutating func handle(_ key: Key, _ o: Options) -> (Bool, [Action]) {
        switch key {
        case .backspace:
            if buffer.isEmpty { return (false, []) }
            buffer.removeLast()
            return (true, [.mark(preview(o))])
        case .escape:
            if buffer.isEmpty { return (false, []) }
            let raw = buffer
            buffer = ""
            return (true, [.commit(raw)])
        case .enter:
            if buffer.isEmpty { return (false, []) }
            return (true, flushActions(o))
        case .space, .other:
            return (false, flushActions(o))
        case .char(let c):
            if Qalam.accepts(c, buffer: buffer) {
                buffer.append(c)
                return (true, [.mark(preview(o))])
            }
            var actions = flushActions(o)
            if let p = Qalam.punctuation(c, o) {
                actions.append(.commit(p))
                return (true, actions)
            }
            return (false, actions)
        }
    }

    public mutating func flushActions(_ o: Options) -> [Action] {
        if buffer.isEmpty { return [] }
        let out = Qalam.word(buffer, o)
        buffer = ""
        return [.commit(out)]
    }

    func preview(_ o: Options) -> String {
        if buffer.isEmpty { return "" }
        let a = Qalam.word(buffer, o)
        return a.isEmpty ? buffer : a
    }
}
