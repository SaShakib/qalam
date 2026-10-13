import SwiftUI
import QalamEngine

/// First launch: a few short steps (keyboard, style, harakat and sukūn, typing helpers, shortcuts,
/// fonts, updates), then the optional 20-word practice. Everything here is also in Settings.
struct SetupView: View {
    @EnvironmentObject var app: AppState

    enum Step: Int, CaseIterable, Identifiable {
        case welcome, keyboard, style, marks, typing, shortcuts, fonts, updates, practice
        var id: Int { rawValue }
        var title: String {
            switch self {
            case .welcome: return "Welcome"
            case .keyboard: return "Keyboard"
            case .style: return "Style"
            case .marks: return "Harakat & sukūn"
            case .typing: return "While typing"
            case .shortcuts: return "Shortcuts"
            case .fonts: return "Fonts"
            case .updates: return "Updates"
            case .practice: return "Practice"
            }
        }
    }

    @State private var step: Step = .welcome

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            ScrollView {
                VStack(alignment: .leading, spacing: 18) {
                    page
                }
                .padding(28)
                .frame(maxWidth: 820, alignment: .leading)
                .frame(maxWidth: .infinity)
            }
            Divider()
            footer
        }
    }

    private var header: some View {
        HStack(spacing: 16) {
            Text("قَلَم").font(app.arabic(40)).foregroundStyle(Color.accentColor)
            VStack(alignment: .leading, spacing: 2) {
                Text("Set up Qalam").font(.title2.bold())
                Text("Step \(step.rawValue + 1) of \(Step.allCases.count): \(step.title). Change any of this later in Settings.")
                    .foregroundStyle(.secondary)
            }
            Spacer()
            HStack(spacing: 6) {
                ForEach(Step.allCases) { s in
                    Circle()
                        .fill(s.rawValue <= step.rawValue ? Color.accentColor : Color.secondary.opacity(0.25))
                        .frame(width: 9, height: 9)
                        .onTapGesture { step = s }
                        .help(s.title)
                }
            }
        }
        .padding(.horizontal, 24).padding(.vertical, 14)
    }

    private var footer: some View {
        HStack {
            Button("Skip setup") { app.finishSetup(practice: false) }
            Spacer()
            if step != .welcome {
                Button("Back") { go(-1) }.controlSize(.large)
            }
            if step == .practice {
                Button("Finish without practice") { app.finishSetup(practice: false) }.controlSize(.large)
                Button("Start the practice") { app.finishSetup(practice: true) }
                    .buttonStyle(.borderedProminent).controlSize(.large)
                    .keyboardShortcut(.defaultAction)
            } else {
                Button("Next") { go(1) }
                    .buttonStyle(.borderedProminent).controlSize(.large)
                    .keyboardShortcut(.defaultAction)
            }
        }
        .padding(.horizontal, 24).padding(.vertical, 12)
    }

    private func go(_ by: Int) {
        if let s = Step(rawValue: step.rawValue + by) { step = s }
    }

    @ViewBuilder
    private var page: some View {
        switch step {
        case .welcome: WelcomeStep()
        case .keyboard: KeyboardStep()
        case .style: StyleStep()
        case .marks: MarksStep()
        case .typing: TypingStep()
        case .shortcuts: ShortcutsStep()
        case .fonts: FontsStep()
        case .updates: UpdatesStep()
        case .practice: PracticeStep()
        }
    }
}

// MARK: Steps

private struct WelcomeStep: View {
    @EnvironmentObject var app: AppState
    private let samples = ["kataba", "kitaAbuN", "eilmuN", "saAala", "madrasat'uN", "alshamsu"]

    var body: some View {
        PageHeader(title: "Welcome to Qalam",
                   subtitle: "Type Arabic the way you spell it: each letter, then its haraka. Sukūn and shadda appear by themselves.")
        Grid(alignment: .leading, horizontalSpacing: 24, verticalSpacing: 8) {
            ForEach(samples, id: \.self) { s in
                GridRow {
                    KeyCap(s)
                    Image(systemName: "arrow.right").foregroundStyle(.secondary)
                    Text(app.show(s)).font(app.arabic(26))
                }
            }
        }
        Text("The next steps turn the keyboard on and set how it writes. They take about two minutes.")
            .foregroundStyle(.secondary)
    }
}

private struct KeyboardStep: View {
    @State private var installed = InputSource.isInstalled
    @State private var enabled = InputSource.isEnabled
    private let timer = Timer.publish(every: 2, on: .main, in: .common).autoconnect()

    var body: some View {
        PageHeader(title: "Turn on the Qalam keyboard",
                   subtitle: "Qalam is a keyboard like ABC or Arabic, so it works in every app.")
        StepCard(number: 1, title: "Add Qalam to your keyboards") {
            if !installed {
                Label("The Qalam keyboard isn't installed. Reinstall Qalam (see the website), then come back.",
                      systemImage: "exclamationmark.triangle.fill")
                    .foregroundStyle(.orange)
            } else if enabled {
                Label("Qalam is in your input menu.", systemImage: "checkmark.circle.fill")
                    .foregroundStyle(.green)
            } else {
                HStack {
                    Button("Add Qalam now") { enabled = InputSource.enable(); InputSource.select() }
                        .buttonStyle(.borderedProminent)
                    Button("Open Keyboard Settings…") { InputSource.openKeyboardSettings() }
                }
                Text("Or by hand: System Settings → Keyboard → Text Input → Edit… → + → Arabic → Qalam. If it doesn't appear, log out and back in once.")
                    .font(.callout).foregroundStyle(.secondary)
            }
        }
        StepCard(number: 2, title: "Switch to it") {
            Text("Press 🌐 Globe or Control-Space until ق shows in the menu bar. Press it again for English.")
            if enabled {
                Button("Switch to Qalam now") { InputSource.select() }
            }
        }
        StepCard(number: 3, title: "How a word is typed") {
            Text("While you type, the word is underlined. Space or punctuation finishes it, Backspace removes the last English key, and Esc gives back the English letters.")
                .foregroundStyle(.secondary)
        }
        .onReceive(timer) { _ in
            guard !enabled else { return }
            installed = InputSource.isInstalled
            enabled = InputSource.isEnabled
        }
    }
}

private struct StyleStep: View {
    @EnvironmentObject var app: AppState
    private let sample = "alHamdu lillaAhi rabbi alea^lamiyna"

    var body: some View {
        PageHeader(title: "Writing style", subtitle: "Everyday is standard modern spelling. Qur'an style adds ٱ, dagger alif, maddah and the mushaf spelling.")
        HStack(alignment: .top, spacing: 14) {
            ChoiceCard(title: "Everyday", detail: "Books, messages, study notes", selected: app.options.style == .everyday,
                       arabic: app.preview(sample) { $0.style = .everyday }) { app.options.style = .everyday }
            ChoiceCard(title: "Qur'an (Uthmani)", detail: "Āyāt as in the mushaf", selected: app.options.style == .quran,
                       arabic: app.preview(sample) { $0.style = .quran }, quran: true) { app.options.style = .quran }
        }
        Toggle("Madinah-mushaf sukūn in Qur'an style (قۡ instead of قْ, needs a mushaf font)", isOn: $app.options.quranSmallSukun)
    }
}

private struct MarksStep: View {
    @EnvironmentObject var app: AppState

    var body: some View {
        PageHeader(title: "Harakat and sukūn", subtitle: "How many marks Qalam writes. The options panel and shortcuts can switch them while you type.")
        SectionTitle("Harakat")
        HStack(alignment: .top, spacing: 12) {
            ForEach(HarakatMode.allCases) { mode in
                ChoiceCard(title: mode.title, detail: detail(mode), selected: app.options.harakat == mode,
                           arabic: app.preview("maktabuN qul") { $0.harakat = mode }) { app.options.harakat = mode }
            }
        }
        SectionTitle("Sukūn")
        HStack(alignment: .top, spacing: 12) {
            ForEach(SukunMode.allCases) { mode in
                ChoiceCard(title: mode.title, detail: detail(mode), selected: app.options.sukun == mode,
                           arabic: app.preview("maktab qul bayt") { $0.harakat = .full; $0.sukun = mode }) { app.options.sukun = mode }
            }
        }
        Text("Words typed with no vowels come out as bare letters: ktb → كتب.").font(.callout).foregroundStyle(.secondary)
    }

    private func detail(_ m: HarakatMode) -> String {
        switch m {
        case .full: return "Every haraka, sukūn added for you"
        case .asTyped: return "No automatic marks"
        case .none: return "Plain letters, for chat and search"
        }
    }

    private func detail(_ m: SukunMode) -> String {
        switch m {
        case .smart: return "Only at a stop inside a word; o adds one"
        case .full: return "On every stop, like a mushaf"
        case .off: return "No sukūn at all"
        }
    }
}

private struct TypingStep: View {
    @EnvironmentObject var app: AppState
    @State private var showOptions = SharedSettings.showOptions

    var body: some View {
        PageHeader(title: "While you type", subtitle: "Small helpers you can switch off.")
        Form {
            Toggle(isOn: $showOptions) {
                VStack(alignment: .leading, spacing: 2) {
                    Text("Show options while typing")
                    Text("Up to 4 versions of the word under it (another hamza seat, no harakat, small alif). ↑↓ choose, Space puts it in. Qalam remembers what you pick, only on this Mac.")
                        .font(.caption).foregroundStyle(.secondary)
                }
            }
            .onChange(of: showOptions) { SharedSettings.showOptions = $0 }
            Toggle(isOn: $app.options.spellingWords) {
                VStack(alignment: .leading, spacing: 2) {
                    Text("Smart spelling")
                    Text(app.preview("allaAh ha^dhaA dha^lika") { $0.spellingWords = true })
                        .font(app.arabic(18))
                }
            }
            Toggle("Arabic digits (١٢٣ instead of 123)", isOn: $app.options.arabicDigits)
        }
        .formStyle(.grouped)
    }
}

private struct ShortcutsStep: View {
    var body: some View {
        PageHeader(title: "Shortcuts", subtitle: "They work while you type with Qalam. Change any that clash with shortcuts you already use.")
        Form {
            ForEach(Shortcut.Action.allCases) { action in
                LabeledContent(action.title) { ShortcutRecorder(action: action) }
            }
            Text("Click a shortcut, then press the new keys. It must use Control or Option. The buttons in the options panel always work too.")
                .font(.caption).foregroundStyle(.secondary)
        }
        .formStyle(.grouped)
    }
}

private struct FontsStep: View {
    @EnvironmentObject var app: AppState

    var body: some View {
        PageHeader(title: "Fonts", subtitle: "The fonts this app shows Arabic in. In other apps, pick the font in that app.")
        Form {
            FontPicker(title: "Arabic font", selection: $app.fontName, installed: app.installedFonts, kfgqpc: app.kfgqpcFamily)
            Text(app.show("AanaA AuHibbu allugat'a alearabiyyat'a"))
                .font(app.arabic(30))
                .frame(maxWidth: .infinity, alignment: .trailing)
            FontPicker(title: "Qur'an font", selection: $app.quranFontName, installed: app.installedFonts, kfgqpc: app.kfgqpcFamily)
            Text(app.show("alHamdu lillaAhi rabbi alea^lamiyna", style: .quran))
                .font(app.arabic(30, quran: true))
                .frame(maxWidth: .infinity, alignment: .trailing)
            Slider(value: $app.fontScale, in: 0.8...1.6) { Text("Arabic size") }
            InstallFontsRow()
            if app.kfgqpcFamily == nil {
                HStack {
                    Text("KFGQPC Hafs (the Madinah mushaf font) isn't installed. Until it is, Qur'an text uses Scheherazade New.")
                        .font(.callout).foregroundStyle(.secondary)
                    Spacer()
                    Link("Get KFGQPC Hafs…", destination: ArabicFonts.kfgqpcDownload)
                }
            }
        }
        .formStyle(.grouped)
        .onAppear { app.scanFonts() }
    }
}

private struct UpdatesStep: View {
    @EnvironmentObject var updater: Updater

    var body: some View {
        PageHeader(title: "Updates", subtitle: "Qalam checks GitHub for a new version when it starts and once a day.")
        Form {
            Toggle("Check for updates automatically", isOn: $updater.autoCheck)
            Toggle("Install updates automatically", isOn: $updater.autoInstall)
                .disabled(updater.method != .perUser)
            if updater.method == .homebrew {
                Text("Installed with Homebrew: updates run `brew upgrade --cask qalam` when you click Install.")
                    .font(.caption).foregroundStyle(.secondary)
            } else if updater.method == .pkg {
                Text("Installed with the .pkg: updates ask for your password once, then install by themselves.")
                    .font(.caption).foregroundStyle(.secondary)
            }
            Text("If the keyboard was on before an update, it stays on afterwards.")
                .font(.caption).foregroundStyle(.secondary)
        }
        .formStyle(.grouped)
    }
}

private struct PracticeStep: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Image(systemName: "checkmark.seal.fill").font(.system(size: 54)).foregroundStyle(.green)
            PageHeader(title: "You're ready",
                       subtitle: "An optional 20-word practice teaches every special key (about 5 minutes). You can skip any word, or come back to it from Practice later.")
            Text("Switch to Qalam with 🌐 or Control-Space and type Arabic anywhere.")
        }
    }
}

// MARK: Pieces

/// One choice among a few, with how it looks in Arabic.
private struct ChoiceCard: View {
    @EnvironmentObject var app: AppState
    let title: String
    let detail: String
    let selected: Bool
    let arabic: String
    var quran = false
    let pick: () -> Void

    var body: some View {
        Button(action: pick) {
            VStack(alignment: .leading, spacing: 6) {
                HStack {
                    Image(systemName: selected ? "largecircle.fill.circle" : "circle")
                        .foregroundStyle(selected ? Color.accentColor : Color.secondary)
                    Text(title).font(.headline)
                }
                Text(detail).font(.callout).foregroundStyle(.secondary)
                Text(arabic)
                    .font(app.arabic(24, quran: quran))
                    .frame(maxWidth: .infinity, alignment: .trailing)
            }
            .padding(14)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(RoundedRectangle(cornerRadius: 12).fill(selected ? Color.accentColor.opacity(0.10) : Color.secondary.opacity(0.06)))
            .overlay(RoundedRectangle(cornerRadius: 12).stroke(selected ? Color.accentColor : Color.secondary.opacity(0.2), lineWidth: selected ? 2 : 1))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }
}
