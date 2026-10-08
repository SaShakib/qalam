import Cocoa
import InputMethodKit

import QalamEngine

// `QalamInput --panel-snapshot out.png word`: draw the options panel to an image and quit (layout check).
if let i = CommandLine.arguments.firstIndex(of: "--panel-snapshot"), i + 2 < CommandLine.arguments.count {
    _ = NSApplication.shared
    let word = CommandLine.arguments[i + 2]
    let panel = OptionsPanel.shared
    panel.show(Qalam.candidates(word, Options()), selected: 1, sukunOff: false, plain: false,
               near: NSRect(x: 300, y: 600, width: 1, height: 20))
    RunLoop.current.run(until: Date().addingTimeInterval(0.5))
    if let view = panel.snapshotView, let rep = view.bitmapImageRepForCachingDisplay(in: view.bounds) {
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
