import SwiftUI
import QalamEngine

struct WelcomeView: View {
    @EnvironmentObject var app: AppState
    @State private var installed = InputSource.isInstalled
    @State private var enabled = InputSource.isEnabled
    private let timer = Timer.publish(every: 2, on: .main, in: .common).autoconnect()

    private let samples = ["kataba", "kitaAbuN", "eilmuN", "saAala", "madrasat'uN", "alshamsu"]

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 22) {
                HStack(alignment: .center, spacing: 20) {
                    Text("قَلَم")
                        .font(app.arabic(64))
                        .foregroundStyle(Color.accentColor)
                    VStack(alignment: .leading, spacing: 6) {
                        Text("Welcome to Qalam").font(.largeTitle.bold())
                        Text("Type Arabic the way you spell it: each letter, then its haraka. Sukūn and shadda appear by themselves.")
                            .foregroundStyle(.secondary)
                    }
                }

                StepCard(number: 1, title: "Add Qalam to your keyboards") {
                    if !installed {
                        Label("The Qalam keyboard is not installed yet. Run `make install` in the Qalam folder.",
                              systemImage: "exclamationmark.triangle.fill")
                            .foregroundStyle(.orange)
                    } else if enabled {
                        Label("Qalam is in your input menu.", systemImage: "checkmark.circle.fill")
                            .foregroundStyle(.green)
                    } else {
                        Text("Click the button, or add it yourself: System Settings → Keyboard → Text Input → Edit… → + → Arabic → Qalam.")
                            .foregroundStyle(.secondary)
                        HStack {
                            Button("Add Qalam now") { enabled = InputSource.enable() }
                                .buttonStyle(.borderedProminent)
                            Button("Open Keyboard Settings…") { InputSource.openKeyboardSettings() }
                        }
                        Text("If Qalam doesn't appear, log out and back in once. macOS sometimes needs this to find a new keyboard.")
                            .font(.callout).foregroundStyle(.secondary)
                    }
                }

                StepCard(number: 2, title: "Switch to it") {
                    Text("Press the 🌐 Globe key, or Control-Space, until the ق icon shows in the menu bar. Press it again to go back to English.")
                    if enabled {
                        Button("Switch to Qalam now") { InputSource.select() }
                    }
                }

                StepCard(number: 3, title: "Type a word, then press Space") {
                    Grid(alignment: .leading, horizontalSpacing: 24, verticalSpacing: 8) {
                        ForEach(samples, id: \.self) { s in
                            GridRow {
                                KeyCap(s)
                                Image(systemName: "arrow.right").foregroundStyle(.secondary)
                                Text(app.show(s)).font(app.arabic(26))
                            }
                        }
                    }
                    Text("While you type, the word is underlined. Space or punctuation finishes it. Backspace removes the last English key. Esc gives back the English letters.")
                        .font(.callout).foregroundStyle(.secondary)
                }

                StepCard(number: 4, title: "Practise") {
                    Text("Start with the 20 starter words, then the word groups. Each item shows exactly what to type.")
                    HStack {
                        Button("Practise the 20 starter words →") {
                            app.practise(.starter)
                        }
                        .buttonStyle(.borderedProminent)
                        .controlSize(.large)
                        Button("See the letter table") { app.page = .letters }
                        .controlSize(.large)
                    }
                }
            }
            .padding(28)
            .frame(maxWidth: 860, alignment: .leading)
        }
        .onReceive(timer) { _ in
            guard !enabled else { return }
            installed = InputSource.isInstalled
            enabled = InputSource.isEnabled
        }
    }
}

struct StepCard<Content: View>: View {
    let number: Int
    let title: String
    @ViewBuilder let content: Content

    var body: some View {
        HStack(alignment: .top, spacing: 16) {
            Text("\(number)")
                .font(.title2.bold())
                .foregroundStyle(.white)
                .frame(width: 36, height: 36)
                .background(Circle().fill(Color.accentColor))
            VStack(alignment: .leading, spacing: 10) {
                Text(title).font(.title3.bold())
                content
            }
            Spacer(minLength: 0)
        }
        .padding(18)
        .background(RoundedRectangle(cornerRadius: 12).fill(Color.secondary.opacity(0.07)))
    }
}
