import SwiftUI
import QalamEngine

@main
struct QalamMain: App {
    @StateObject private var app = AppState()
    @StateObject private var updater = Updater()

    init() {
        Snapshot.runIfRequested()
        InputSource.restoreIfNeeded()
    }

    var body: some Scene {
        WindowGroup("Qalam") {
            ContentView()
                .environmentObject(app)
                .environmentObject(updater)
                .onAppear { updater.start() }
                .frame(minWidth: 920, minHeight: 640)
        }
        .defaultSize(width: 1100, height: 780)
        .commands {
            CommandGroup(replacing: .newItem) {}
        }
    }
}

struct ContentView: View {
    @EnvironmentObject var app: AppState

    var body: some View {
        VStack(spacing: 0) {
            UpdateBanner()
            if app.firstPracticeDone { mainView } else { FirstRunView() }
        }
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didBecomeActiveNotification)) { _ in
            // "Change shortcuts…" in the keyboard's menu opens Settings.
            if let p = SharedSettings.takeRequestedPage().flatMap(Page.init(rawValue:)) { app.page = p }
        }
    }

    private var mainView: some View {
        NavigationSplitView {
            List(selection: $app.page) {
                ForEach(Page.allCases) { p in
                    Label(p.title, systemImage: p.icon).tag(p)
                }
            }
            .navigationSplitViewColumnWidth(min: 170, ideal: 190)
        } detail: {
            switch app.page ?? .letters {
            case .welcome: WelcomeView()
            case .letters: LettersView()
            case .words: WordsView()
            case .practice: PracticeView()
            case .write: WriteView()
            case .settings: SettingsView()
            }
        }
    }
}

/// Shown above everything when an update is available, installing, or failed.
struct UpdateBanner: View {
    @EnvironmentObject var updater: Updater

    var body: some View {
        switch updater.status {
        case .available(let v, _):
            bar("arrow.down.circle.fill", "Qalam \(v) is available (you have \(updater.current)).", .accentColor) {
                Button(updater.method == .pkg ? "Download installer" : "Install update") { Task { await updater.install() } }
                    .buttonStyle(.borderedProminent)
            }
        case .working(let text):
            bar("arrow.triangle.2.circlepath", text, .accentColor) { ProgressView().controlSize(.small) }
        case .installed(let v):
            bar("checkmark.circle.fill", "Updated to \(v). Restarting…", .green) { EmptyView() }
        case .failed(let msg):
            bar("exclamationmark.triangle.fill", "Update failed: \(msg)", .orange) {
                Button("Try again") { Task { await updater.check(userInitiated: true) } }
            }
        default:
            EmptyView()
        }
    }

    private func bar<Trailing: View>(_ icon: String, _ text: String, _ color: Color,
                                     @ViewBuilder trailing: () -> Trailing) -> some View {
        HStack(spacing: 10) {
            Image(systemName: icon).foregroundStyle(color)
            Text(text).lineLimit(2)
            Spacer()
            trailing()
        }
        .padding(.horizontal, 16).padding(.vertical, 8)
        .background(color.opacity(0.12))
    }
}
