import SwiftUI
import AppKit

/// `Qalam --snapshot <dir>` draws every page to a PNG off-screen and quits (used to check the layout).
enum Snapshot {
    static func runIfRequested() {
        let args = CommandLine.arguments
        guard let i = args.firstIndex(of: "--snapshot"), i + 1 < args.count else { return }
        let dir = URL(fileURLWithPath: args[i + 1])
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        _ = NSApplication.shared
        let size = NSSize(width: 1100, height: 780)
        for page in [nil] + Page.allCases.map(Optional.some) {
            let state = AppState()
            state.firstPracticeDone = page != nil
            state.page = page ?? .letters
            let root = ContentView().environmentObject(state)
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
