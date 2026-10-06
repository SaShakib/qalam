import SwiftUI

@main
struct QalamMain: App {
    @StateObject private var app = AppState()

    init() {
        Snapshot.runIfRequested()
    }

    var body: some Scene {
        WindowGroup("Qalam") {
            ContentView()
                .environmentObject(app)
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
        if app.firstPracticeDone { mainView } else { FirstRunView() }
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
