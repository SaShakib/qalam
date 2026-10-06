import AppKit
import Foundation

/// Checks GitHub Releases for a newer Qalam and installs it.
/// How it installs depends on how Qalam was installed:
///   • install.sh / per-user copy → downloads the zip and replaces both bundles in place
///   • Homebrew                   → runs `brew upgrade --cask qalam`
///   • .pkg in /Applications      → downloads the new .pkg and opens it in Installer
@MainActor
final class Updater: ObservableObject {
    static let repo = "SaShakib/qalam"

    enum Status: Equatable {
        case idle
        case checking
        case upToDate
        case available(version: String, notes: String)
        case working(String)
        case installed(String)
        case failed(String)
    }

    enum Method { case perUser, homebrew, pkg }

    @Published var status: Status = .idle
    @Published var autoCheck: Bool { didSet { UserDefaults.standard.set(autoCheck, forKey: "autoCheckUpdates") } }
    @Published var autoInstall: Bool { didSet { UserDefaults.standard.set(autoInstall, forKey: "autoInstallUpdates") } }
    @Published var lastChecked: Date? {
        didSet { UserDefaults.standard.set(lastChecked, forKey: "lastUpdateCheck") }
    }

    let current: String = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "0"
    private var latest: Release?
    private var timer: Timer?

    init() {
        let d = UserDefaults.standard
        autoCheck = d.object(forKey: "autoCheckUpdates") as? Bool ?? true
        autoInstall = d.object(forKey: "autoInstallUpdates") as? Bool ?? true
        lastChecked = d.object(forKey: "lastUpdateCheck") as? Date
    }

    /// Called once at launch: check now (if allowed), then once a day while the app is open.
    func start() {
        guard timer == nil else { return }
        if autoCheck { Task { await check(userInitiated: false) } }
        timer = Timer.scheduledTimer(withTimeInterval: 24 * 3600, repeats: true) { [weak self] _ in
            Task { @MainActor in
                guard let self, self.autoCheck else { return }
                await self.check(userInitiated: false)
            }
        }
    }

    var method: Method {
        let fm = FileManager.default
        if fm.fileExists(atPath: "/opt/homebrew/Caskroom/qalam") || fm.fileExists(atPath: "/usr/local/Caskroom/qalam") {
            return .homebrew
        }
        if Bundle.main.bundlePath.hasPrefix("/Applications/")
            && fm.fileExists(atPath: "/Library/Input Methods/QalamInput.app") {
            return .pkg
        }
        return .perUser
    }

    // MARK: Checking

    struct Release: Decodable {
        let tag_name: String
        let body: String?
        let html_url: String
        let assets: [Asset]
        struct Asset: Decodable {
            let name: String
            let browser_download_url: URL
        }
        var version: String { tag_name.hasPrefix("v") ? String(tag_name.dropFirst()) : tag_name }
        func asset(_ test: (String) -> Bool) -> URL? { assets.first { test($0.name) }?.browser_download_url }
    }

    func check(userInitiated: Bool) async {
        if case .working = status { return }
        status = .checking
        do {
            var req = URLRequest(url: URL(string: "https://api.github.com/repos/\(Self.repo)/releases/latest")!)
            req.setValue("application/vnd.github+json", forHTTPHeaderField: "Accept")
            req.setValue("Qalam/\(current)", forHTTPHeaderField: "User-Agent")
            req.timeoutInterval = 20
            let (data, response) = try await URLSession.shared.data(for: req)
            guard (response as? HTTPURLResponse)?.statusCode == 200 else { throw UpdateError("GitHub did not answer (try again later).") }
            let release = try JSONDecoder().decode(Release.self, from: data)
            lastChecked = Date()
            latest = release
            if Self.isNewer(release.version, than: current) {
                status = .available(version: release.version, notes: release.body ?? "")
                if !userInitiated && autoInstall && method == .perUser {
                    await install()
                }
            } else {
                status = userInitiated ? .upToDate : .idle
            }
        } catch {
            status = userInitiated ? .failed(error.localizedDescription) : .idle
        }
    }

    static func isNewer(_ a: String, than b: String) -> Bool {
        let pa = a.split(separator: ".").map { Int($0) ?? 0 }
        let pb = b.split(separator: ".").map { Int($0) ?? 0 }
        for i in 0..<max(pa.count, pb.count) {
            let x = i < pa.count ? pa[i] : 0, y = i < pb.count ? pb[i] : 0
            if x != y { return x > y }
        }
        return false
    }

    // MARK: Installing

    func install() async {
        guard let release = latest else { return }
        do {
            switch method {
            case .perUser: try await installZip(release)
            case .homebrew: try await installBrew()
            case .pkg: try await openPkg(release)
            }
        } catch {
            status = .failed(error.localizedDescription)
        }
    }

    private func installZip(_ release: Release) async throws {
        guard let url = release.asset({ $0 == "Qalam-mac.zip" }) else { throw UpdateError("The release has no Mac download.") }
        status = .working("Downloading Qalam \(release.version)…")
        let (tmpFile, _) = try await URLSession.shared.download(from: url)
        let work = FileManager.default.temporaryDirectory.appendingPathComponent("qalam-update-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: work, withIntermediateDirectories: true)
        defer { try? FileManager.default.removeItem(at: work) }
        let zip = work.appendingPathComponent("Qalam-mac.zip")
        try FileManager.default.moveItem(at: tmpFile, to: zip)

        status = .working("Installing…")
        try run("/usr/bin/ditto", ["-x", "-k", zip.path, work.appendingPathComponent("x").path])
        let newApp = work.appendingPathComponent("x/Qalam.app")
        let newIME = work.appendingPathComponent("x/QalamInput.app")
        guard FileManager.default.fileExists(atPath: newApp.path),
              FileManager.default.fileExists(atPath: newIME.path) else { throw UpdateError("The download looks incomplete.") }

        let appDest = Bundle.main.bundleURL
        let imeDest = InputSource.bundleURL
        let wasEnabled = InputSource.isEnabled || InputSource.rememberedEnabled

        _ = try? run("/usr/bin/killall", ["QalamInput"])
        try replace(imeDest, with: newIME)
        try replace(appDest, with: newApp)
        _ = try? run("/usr/bin/xattr", ["-dr", "com.apple.quarantine", appDest.path, imeDest.path])
        if wasEnabled { InputSource.enable() }

        status = .installed(release.version)
        relaunch()
    }

    private func replace(_ dest: URL, with new: URL) throws {
        let fm = FileManager.default
        try fm.createDirectory(at: dest.deletingLastPathComponent(), withIntermediateDirectories: true)
        let old = dest.deletingLastPathComponent().appendingPathComponent(".\(dest.lastPathComponent).old")
        try? fm.removeItem(at: old)
        if fm.fileExists(atPath: dest.path) { try fm.moveItem(at: dest, to: old) }
        do {
            try run("/usr/bin/ditto", [new.path, dest.path])
            try? fm.removeItem(at: old)
        } catch {
            if fm.fileExists(atPath: old.path) { try? fm.moveItem(at: old, to: dest) }
            throw error
        }
    }

    private func installBrew() async throws {
        let brew = ["/opt/homebrew/bin/brew", "/usr/local/bin/brew"].first { FileManager.default.isExecutableFile(atPath: $0) }
        guard let brew else { throw UpdateError("Homebrew was not found. Run: brew upgrade --cask qalam") }
        status = .working("Updating with Homebrew… (this can take a minute)")
        let wasEnabled = InputSource.isEnabled || InputSource.rememberedEnabled
        try await Task.detached { try Self.runStatic(brew, ["upgrade", "--cask", "qalam"]) }.value
        if wasEnabled { InputSource.enable() }
        status = .installed(latest?.version ?? "")
        relaunch()
    }

    private func openPkg(_ release: Release) async throws {
        guard let url = release.asset({ $0.hasSuffix(".pkg") }) else { throw UpdateError("The release has no installer package.") }
        status = .working("Downloading the installer…")
        let (tmpFile, _) = try await URLSession.shared.download(from: url)
        let dest = FileManager.default.temporaryDirectory.appendingPathComponent(url.lastPathComponent)
        try? FileManager.default.removeItem(at: dest)
        try FileManager.default.moveItem(at: tmpFile, to: dest)
        NSWorkspace.shared.open(dest)
        status = .working("Finish the update in the Installer window.")
    }

    private func relaunch() {
        let path = Bundle.main.bundlePath
        let p = Process()
        p.executableURL = URL(fileURLWithPath: "/bin/sh")
        p.arguments = ["-c", "sleep 1; /usr/bin/open \"\(path)\""]
        try? p.run()
        NSApp.terminate(nil)
    }

    @discardableResult
    private func run(_ tool: String, _ args: [String]) throws -> String { try Self.runStatic(tool, args) }

    @discardableResult
    nonisolated static func runStatic(_ tool: String, _ args: [String]) throws -> String {
        let p = Process()
        p.executableURL = URL(fileURLWithPath: tool)
        p.arguments = args
        let out = Pipe()
        p.standardOutput = out
        p.standardError = out
        try p.run()
        p.waitUntilExit()
        let text = String(data: out.fileHandleForReading.readDataToEndOfFile(), encoding: .utf8) ?? ""
        if p.terminationStatus != 0 {
            throw UpdateError("\((tool as NSString).lastPathComponent) failed: \(text.suffix(300))")
        }
        return text
    }
}

struct UpdateError: LocalizedError {
    let message: String
    init(_ m: String) { message = m }
    var errorDescription: String? { message }
}
