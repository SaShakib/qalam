import AppKit
import Carbon
import ServiceManagement
import QalamEngine

/// Qalam keeps running in the menu bar after its window closes, and starts at login (hidden), so
/// ⌃⇧Q opens the Qur'an box in any app, whatever keyboard is on.
/// `Qalam --background` (how the login agent starts it): no window, just the menu bar item.
final class AppDelegate: NSObject, NSApplicationDelegate {
    private var status: NSStatusItem?
    static let agentPlist = "com.qalam.background.plist"

    func applicationDidFinishLaunching(_ notification: Notification) {
        let background = CommandLine.arguments.contains("--background")
        NSLog("Qalam: launched\(background ? " in the background" : "")")
        // One Qalam at a time: the login agent's copy steps aside, and a normal launch hands over to the
        // copy already in the menu bar (it shows its window).
        let others = NSRunningApplication.runningApplications(withBundleIdentifier: Bundle.main.bundleIdentifier ?? "")
            .filter { $0.processIdentifier != getpid() }
        if !others.isEmpty {
            if !background {
                DistributedNotificationCenter.default().postNotificationName(.init("com.qalam.app.show"), object: nil,
                                                                             userInfo: nil, deliverImmediately: true)
            }
            exit(0)
        }
        DistributedNotificationCenter.default().addObserver(forName: .init("com.qalam.app.show"), object: nil, queue: .main) { [weak self] _ in
            self?.openWindow()
        }
        if background {
            NSApp.setActivationPolicy(.accessory)
            DispatchQueue.main.async { NSApp.windows.filter { $0.canBecomeMain }.forEach { $0.close() } }
        }
        QuranBox.shared.start()
        if Self.runInBackground { Self.setLoginAgent(true) }
        updateStatusItem()
        // The keyboard asked for the box before Qalam was running.
        if SharedSettings.takeRequestedPage(matching: "quranbox") { QuranBox.shared.open() }

        let nc = NotificationCenter.default
        nc.addObserver(forName: NSWindow.willCloseNotification, object: nil, queue: .main) { [weak self] _ in
            DispatchQueue.main.async { self?.windowsChanged() }
        }
        nc.addObserver(forName: NSApplication.didResignActiveNotification, object: nil, queue: .main) { _ in
            QuranHotKey.shared.register()   // a shortcut changed in Settings
        }
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool { !Self.runInBackground }

    func applicationShouldHandleReopen(_ sender: NSApplication, hasVisibleWindows flag: Bool) -> Bool {
        if !flag { NSApp.setActivationPolicy(.regular) }
        return true   // SwiftUI opens a window
    }

    /// No main window left: no Dock icon, just the menu bar item.
    private func windowsChanged() {
        guard Self.runInBackground else { return }
        let open = NSApp.windows.contains { $0.isVisible && $0.canBecomeMain && !($0 is NSPanel) }
        if !open { NSApp.setActivationPolicy(.accessory) }
    }

    // MARK: Menu bar item

    func updateStatusItem() {
        guard Self.runInBackground else {
            if let status { NSStatusBar.system.removeStatusItem(status) }
            status = nil
            return
        }
        let item = status ?? NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        item.button?.title = "ق"
        item.button?.toolTip = "Qalam"
        let menu = NSMenu()
        let key = SharedSettings.shortcut(for: .quran)?.display ?? ""
        menu.addItem(withTitle: "Insert from the Qur'an…  \(key)", action: #selector(openBox), keyEquivalent: "").target = self
        menu.addItem(withTitle: "Open Qalam", action: #selector(openWindow), keyEquivalent: "").target = self
        menu.addItem(.separator())
        menu.addItem(withTitle: "Quit Qalam", action: #selector(NSApplication.terminate(_:)), keyEquivalent: "")
        item.menu = menu
        status = item
    }

    @objc private func openBox() { QuranBox.shared.open() }

    @objc func openWindow() {
        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)
        if let w = NSApp.windows.first(where: { $0.canBecomeMain && !($0 is NSPanel) }) {
            w.makeKeyAndOrderFront(nil)
        } else {
            // SwiftUI opens a new window for a "reopen" Apple event; send one to ourselves.
            let me = NSAppleEventDescriptor(processIdentifier: getpid())
            let event = NSAppleEventDescriptor.appleEvent(withEventClass: AEEventClass(kCoreEventClass),
                                                          eventID: AEEventID(kAEReopenApplication), targetDescriptor: me,
                                                          returnID: AEReturnID(kAutoGenerateReturnID),
                                                          transactionID: AETransactionID(kAnyTransactionID))
            _ = try? event.sendEvent(options: [.noReply], timeout: 2)
        }
    }

    // MARK: Setting

    /// "Keep Qalam running for ⌃⇧Q" (Settings), on by default.
    static var runInBackground: Bool {
        get { UserDefaults.standard.object(forKey: "runInBackground") as? Bool ?? true }
        set {
            UserDefaults.standard.set(newValue, forKey: "runInBackground")
            setLoginAgent(newValue)
            (NSApp.delegate as? AppDelegate)?.updateStatusItem()
        }
    }

    /// Starts Qalam hidden at login (a launch agent inside Qalam.app). macOS may ask once in Login Items.
    static func setLoginAgent(_ on: Bool) {
        let service = SMAppService.agent(plistName: agentPlist)
        do {
            if on, service.status != .enabled { try service.register() }
            if !on, service.status == .enabled { try service.unregister() }
        } catch {
            NSLog("Qalam login agent: \(error.localizedDescription)")
        }
    }

    static var loginAgentNeedsApproval: Bool { SMAppService.agent(plistName: agentPlist).status == .requiresApproval }
}

extension SharedSettings {
    /// True (and forgotten) when the keyboard asked for `page`.
    static func takeRequestedPage(matching page: String) -> Bool {
        guard let p = takeRequestedPage() else { return false }
        if p == page { return true }
        requestPage(p)   // someone else's request: leave it for the window
        return false
    }
}
