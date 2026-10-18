import SwiftUI
import ServiceManagement
import QalamEngine

/// What the Qur'an box (⌃⇧Q) needs from macOS, with a live status for each: in the setup steps,
/// once as a sheet for people who set Qalam up before the box existed, and in Settings.
struct QuranAccessView: View {
    @State private var running = AppDelegate.runInBackground
    @State private var loginStatus = SMAppService.agent(plistName: AppDelegate.agentPlist).status
    @State private var canPaste = QuranInserter.canPaste
    private let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()
    private var key: String { SharedSettings.shortcut(for: .quran)?.display ?? "the Qur'an shortcut" }

    var body: some View {
        PageHeader(title: "Qur'an box (\(key))",
                   subtitle: "Press \(key) in any app to search the Qur'an and insert āyāt with full harakat. Two permissions make it work everywhere.")

        StepCard(number: 1, title: "Start Qalam at login") {
            Text("Qalam waits in the menu bar (ق) so \(key) works in every app, with any keyboard, even after a restart. It uses almost no memory.")
                .foregroundStyle(.secondary)
            switch (running, loginStatus) {
            case (true, .enabled):
                done("Qalam starts at login.")
            case (true, .requiresApproval):
                Label("macOS needs your OK: in Login Items, switch on Qalam under “Allow in the Background”.",
                      systemImage: "exclamationmark.circle.fill").foregroundStyle(.orange)
                Button("Open Login Items…") { SMAppService.openSystemSettingsLoginItems() }
                    .buttonStyle(.borderedProminent)
            default:
                Button("Start Qalam at login") {
                    AppDelegate.runInBackground = true
                    refresh()
                }
                .buttonStyle(.borderedProminent)
            }
        }

        StepCard(number: 2, title: "Allow pasting (Accessibility)") {
            Text("With the Qalam keyboard on, the āyāt are typed in directly and nothing else is needed. With another keyboard on (ABC, Lekho…), macOS lets Qalam paste them only with this permission. Without it they are copied and you press ⌘V.")
                .foregroundStyle(.secondary)
            if canPaste {
                done("Qalam can paste into any app.")
            } else {
                HStack {
                    Button("Allow…") { QuranInserter.askForAccessibility() }
                        .buttonStyle(.borderedProminent)
                    Text("System Settings opens at Accessibility: switch on Qalam. If it was on before an update, Qalam clears the old entry, so switch it on again.")
                        .font(.callout).foregroundStyle(.secondary)
                }
            }
        }

        StepCard(number: 3, title: "Try it") {
            Text("Click in any text field, press \(key), search a word or pick a sūrah and āyāt, then Insert.")
                .foregroundStyle(.secondary)
            Button("Open the Qur'an box now") { QuranBox.shared.open() }
        }
        .onReceive(timer) { _ in refresh() }
    }

    private func refresh() {
        running = AppDelegate.runInBackground
        loginStatus = SMAppService.agent(plistName: AppDelegate.agentPlist).status
        canPaste = QuranInserter.canPaste
    }

    private func done(_ text: String) -> some View {
        Label(text, systemImage: "checkmark.circle.fill").foregroundStyle(.green)
    }
}

/// Shown once to people who set Qalam up before the Qur'an box existed.
struct QuranAccessSheet: View {
    let close: () -> Void
    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            QuranAccessView()
            HStack {
                Spacer()
                Button("Done", action: close).buttonStyle(.borderedProminent).keyboardShortcut(.defaultAction)
            }
        }
        .padding(28)
        .frame(width: 640)
    }
}
