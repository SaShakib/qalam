import Cocoa
import InputMethodKit

import QalamEngine

// `QalamInput --panel-snapshot out.png word`: draw the options panel to an image and quit (layout check).
if let i = CommandLine.arguments.firstIndex(of: "--panel-snapshot"), i + 2 < CommandLine.arguments.count {
    _ = NSApplication.shared
    let word = CommandLine.arguments[i + 2]
    let panel = OptionsPanel.shared
    panel.show(Qalam.candidates(word, Options()), selected: 1, sukunOff: false, plain: false,
               near: NSRect(x: 300, y: 600, width: 1, height: 20),
               sukunKey: SharedSettings.shortcut(for: .sukun)?.display,
               harakatKey: SharedSettings.shortcut(for: .harakat)?.display)
    RunLoop.current.run(until: Date().addingTimeInterval(0.5))
    if let view = panel.snapshotView, let rep = view.bitmapImageRepForCachingDisplay(in: view.bounds) {
        view.cacheDisplay(in: view.bounds, to: rep)
        try? rep.representation(using: .png, properties: [:])?.write(to: URL(fileURLWithPath: CommandLine.arguments[i + 1]))
    }
    exit(0)
}

// `QalamInput --quran-snapshot out.png "query"`: draw the Qur'an box to an image and quit (layout check).
if let i = CommandLine.arguments.firstIndex(of: "--quran-snapshot"), i + 2 < CommandLine.arguments.count {
    _ = NSApplication.shared
    let q = CommandLine.arguments[i + 2]
    let found: QuranLibrary? = QuranStore.shared ?? QuranLibrary(folders: [URL(fileURLWithPath: "quran")])
    guard let lib = found else { print("no Qur'an data"); exit(1) }
    var plain = Options(); plain.harakat = .none
    let rows = lib.suggestions(for: q, arabic: q.contains(where: { $0.isLetter && $0.isASCII }) ? Qalam.text(q, plain) : nil, limit: 6)
    QuranPanel.shared.show(query: q, rows: rows, selected: 0, total: nil, library: lib, mode: SharedSettings.quranInsert,
                           translation: lib.translation(SharedSettings.quranTranslation),
                           near: NSRect(x: 300, y: 700, width: 1, height: 20), key: "⌃⇧Q")
    RunLoop.current.run(until: Date().addingTimeInterval(0.5))
    if let view = QuranPanel.shared.snapshotView, let rep = view.bitmapImageRepForCachingDisplay(in: view.bounds) {
        view.cacheDisplay(in: view.bounds, to: rep)
        try? rep.representation(using: .png, properties: [:])?.write(to: URL(fileURLWithPath: CommandLine.arguments[i + 1]))
    }
    exit(0)
}

// macOS starts this process when Qalam is chosen in the input menu.
let connectionName = Bundle.main.infoDictionary?["InputMethodConnectionName"] as? String
    ?? "com.qalam.inputmethod.Qalam_Connection"
let server = IMKServer(name: connectionName, bundleIdentifier: Bundle.main.bundleIdentifier)
_ = server
NSApplication.shared.run()
