import SwiftUI
import AppKit

/// `Qalam --snapshot <dir>` draws every page to a PNG off-screen and quits (used to check the layout).
@MainActor
enum Snapshot {
    static func runIfRequested() {
        let args = CommandLine.arguments
        // `Qalam --quran-box-snapshot out.png "search words"`: the Qur'an box with that search.
        if let i = args.firstIndex(of: "--quran-box-snapshot"), i + 2 < args.count {
            _ = NSApplication.shared
            ArabicFonts.registerBundled()
            guard let lib = QuranData.library else { print("no Qur'an data"); exit(1) }
            let model = QuranBoxModel(lib: lib)
            model.query = args[i + 2]
            if args.count > i + 3, let n = Int(args[i + 3]) { model.move(n); model.addAfter(); model.addAfter() }
            let size = NSSize(width: 760, height: 620)
            let host = NSHostingView(rootView: QuranBoxView(model: model, insert: {}, cancel: {})
                .frame(width: size.width, height: size.height).background(Color(nsColor: .windowBackgroundColor)))
            host.appearance = NSAppearance(named: .aqua)
            host.frame = NSRect(origin: .zero, size: size)
            let window = NSWindow(contentRect: host.frame, styleMask: [.titled], backing: .buffered, defer: false)
            window.contentView = host
            host.layoutSubtreeIfNeeded()
            RunLoop.current.run(until: Date().addingTimeInterval(0.8))
            if let rep = host.bitmapImageRepForCachingDisplay(in: host.bounds) {
                host.cacheDisplay(in: host.bounds, to: rep)
                try? rep.representation(using: .png, properties: [:])?.write(to: URL(fileURLWithPath: args[i + 1]))
            }
            exit(0)
        }
        guard let i = args.firstIndex(of: "--snapshot"), i + 1 < args.count else { return }
        let dir = URL(fileURLWithPath: args[i + 1])
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        _ = NSApplication.shared
        let size = NSSize(width: 1100, height: 780)
        for page in [nil] + Page.allCases.map(Optional.some) {
            let state = AppState()
            state.firstPracticeDone = page != nil
            state.page = page ?? .letters
            let root = ContentView().environmentObject(state).environmentObject(Updater())
                .frame(width: size.width, height: size.height)
                .background(Color(nsColor: .windowBackgroundColor))
            let host = NSHostingView(rootView: root)
            host.appearance = NSAppearance(named: .aqua)
            host.frame = NSRect(origin: .zero, size: size)
            let window = NSWindow(contentRect: host.frame, styleMask: [.titled], backing: .buffered, defer: false)
            window.contentView = host
            let t0 = Date()
            host.layoutSubtreeIfNeeded()
            host.display()
            print("page \(page?.rawValue ?? "firstrun"): \(Int(Date().timeIntervalSince(t0) * 1000)) ms")
            RunLoop.current.run(until: Date().addingTimeInterval(0.6))
            if let rep = host.bitmapImageRepForCachingDisplay(in: host.bounds) {
                host.cacheDisplay(in: host.bounds, to: rep)
                try? rep.representation(using: .png, properties: [:])?.write(to: dir.appendingPathComponent("\(page?.rawValue ?? "firstrun").png"))
            }
        }
        exit(0)
    }
}
