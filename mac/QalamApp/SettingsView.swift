import SwiftUI
import QalamEngine

struct SettingsView: View {
    @EnvironmentObject var app: AppState
    @EnvironmentObject var updater: Updater
    @State private var enabled = InputSource.isEnabled
    @State private var showOptions = SharedSettings.showOptions
    @State private var forgotten = false

    var body: some View {
        Form {
            Section("Keyboard: applies in every app") {
                Picker("Writing style", selection: $app.options.style) {
                    ForEach(Style.allCases) { Text($0.title).tag($0) }
                }
                Picker("Harakat", selection: $app.options.harakat) {
                    ForEach(HarakatMode.allCases) { Text($0.title).tag($0) }
                }
                Picker("Sukūn", selection: $app.options.sukun) {
                    ForEach(SukunMode.allCases) { Text($0.title).tag($0) }
                }
                Text("Smart: only at a stop inside a word (مَكْتَب، قُل، بَيت); typing o always adds one. Off: none at all (the sukūn shortcut below, or the button in the options panel). Words typed with no vowels come out as bare letters (ktb → كتب).")
                    .font(.caption).foregroundStyle(.secondary)
                Toggle("Show options while typing (↑↓ to choose, like Avro)", isOn: $showOptions)
                    .onChange(of: showOptions) { SharedSettings.showOptions = $0 }
                HStack {
                    Button(forgotten ? "Forgotten ✓" : "Forget what I picked") {
                        SharedSettings.forgetChoices()
                        forgotten = true
                    }
                    Text("Qalam remembers the option you pick for each word, only on this Mac.")
                        .font(.caption).foregroundStyle(.secondary)
                }
                Toggle("Smart spelling words (اللَّه، هَٰذَا، ذَٰلِكَ، الَّذِي …)", isOn: $app.options.spellingWords)
                Toggle("Arabic digits (١٢٣)", isOn: $app.options.arabicDigits)
                Toggle("Qur'an style: Madinah-mushaf sukūn (قۡ instead of قْ)", isOn: $app.options.quranSmallSukun)
                Text(app.show("alHamdu lillaAhi rabbi alea^lamiyna"))
                    .font(app.arabic(26))
                    .frame(maxWidth: .infinity, alignment: .trailing)
            }

            Section("Shortcuts: work while typing with Qalam") {
                ForEach(Shortcut.Action.allCases) { action in
                    LabeledContent(action.title) { ShortcutRecorder(action: action) }
                }
                Text("Click a shortcut, then press the new keys. It must use Control or Option (⌘ shortcuts belong to apps). If a shortcut clashes with one you use in another app, change it here or remove it; the buttons in the options panel always work.")
                    .font(.caption).foregroundStyle(.secondary)
            }

            Section("Fonts in this app") {
                FontPicker(title: "Arabic font", selection: $app.fontName, installed: app.installedFonts, kfgqpc: app.kfgqpcFamily)
                Text(app.show("AanaA AuHibbu allugat'a alearabiyyat'a"))
                    .font(app.arabic(30))
                    .frame(maxWidth: .infinity, alignment: .trailing)
                FontPicker(title: "Qur'an font", selection: $app.quranFontName, installed: app.installedFonts, kfgqpc: app.kfgqpcFamily)
                Text(app.show("alHamdu lillaAhi rabbi alea^lamiyna", style: .quran))
                    .font(app.arabic(30, quran: true))
                    .frame(maxWidth: .infinity, alignment: .trailing)
                Slider(value: $app.fontScale, in: 0.8...1.6) { Text("Arabic size") }
                if app.kfgqpcFamily == nil {
                    HStack {
                        Text("KFGQPC Hafs (the Madinah mushaf font) isn't installed. Install it, then pick it above.")
                            .font(.callout).foregroundStyle(.secondary)
                        Spacer()
                        Link("Get KFGQPC Hafs…", destination: ArabicFonts.kfgqpcDownload)
                    }
                }
                Text("Fonts apply to this app. In other apps, choose the font in that app (these fonts can be installed from the Qalam folder's fonts/ directory).")
                    .font(.caption).foregroundStyle(.secondary)
            }

            Section("Updates") {
                LabeledContent("Version", value: updater.current)
                Toggle("Check for updates automatically", isOn: $updater.autoCheck)
                Toggle("Install updates automatically", isOn: $updater.autoInstall)
                    .disabled(updater.method != .perUser)
                HStack {
                    Button("Check for updates now") { Task { await updater.check(userInitiated: true) } }
                    if case .available = updater.status {
                        Button("Install update") { Task { await updater.install() } }.buttonStyle(.borderedProminent)
                    }
                    Spacer()
                    Text(updateText).foregroundStyle(.secondary)
                }
                if updater.method == .homebrew {
                    Text("Installed with Homebrew: updates run `brew upgrade --cask qalam`.").font(.caption).foregroundStyle(.secondary)
                } else if updater.method == .pkg {
                    Text("Installed with the .pkg: updates download the new installer for you to open.").font(.caption).foregroundStyle(.secondary)
                }
            }

            Section("Keyboard status") {
                LabeledContent("Installed", value: InputSource.isInstalled ? "Yes" : "No: run make install")
                LabeledContent("In the input menu", value: enabled ? "Yes" : "No")
                HStack {
                    if !enabled {
                        Button("Add Qalam now") { enabled = InputSource.enable() }
                    }
                    Button("Open Keyboard Settings…") { InputSource.openKeyboardSettings() }
                    Button("Show welcome again") { app.page = .welcome }
                    Button("Do the 20-word starter practice again") { app.restartFirstPractice() }
                }
            }
        }
        .formStyle(.grouped)
        .onAppear { enabled = InputSource.isEnabled; app.scanFonts() }
    }
}

extension SettingsView {
    var updateText: String {
        switch updater.status {
        case .checking: return "Checking…"
        case .upToDate: return "You have the latest version."
        case .available(let v, _): return "Version \(v) is available."
        case .working(let t): return t
        case .installed(let v): return "Updated to \(v)."
        case .failed(let m): return m
        case .idle:
            if let d = updater.lastChecked {
                return "Last checked " + d.formatted(.relative(presentation: .named))
            }
            return ""
        }
    }
}

struct FontPicker: View {
    let title: String
    @Binding var selection: String
    let installed: [String]
    let kfgqpc: String?

    var body: some View {
        Picker(title, selection: $selection) {
            Section("Included with Qalam") {
                ForEach(ArabicFonts.bundled) { f in
                    Text("\(f.family) (\(f.note))").tag(f.family)
                }
            }
            if let k = kfgqpc {
                Section("Qur'an Complex") { Text(k).tag(k) }
            }
            Section("Installed on this Mac") {
                Text("System font").tag("")
                ForEach(installed.filter { $0 != kfgqpc }, id: \.self) { Text($0).tag($0) }
            }
        }
    }
}

/// Click, then press keys: records a shortcut for the Qalam keyboard.
struct ShortcutRecorder: View {
    let action: Shortcut.Action
    @State private var current: Shortcut?
    @State private var recording = false
    @State private var message = ""
    @State private var monitor: Any?

    var body: some View {
        HStack(spacing: 8) {
            if !message.isEmpty {
                Text(message).font(.caption).foregroundStyle(.orange)
            }
            Button(recording ? "Press keys… (Esc cancels)" : (current?.display ?? "None")) {
                recording ? stop() : start()
            }
            .font(.system(.body, design: .rounded).weight(.medium))
            .frame(minWidth: 150)
            Menu {
                Button("Remove shortcut") { save(nil) }
                Button("Reset to default (\(action.defaultShortcut?.display ?? "none"))") {
                    SharedSettings.resetShortcut(for: action)
                    current = SharedSettings.shortcut(for: action)
                    message = ""
                }
            } label: { Image(systemName: "ellipsis.circle") }
            .menuStyle(.borderlessButton).fixedSize()
        }
        .onAppear { current = SharedSettings.shortcut(for: action) }
        .onDisappear { stop() }
    }

    private func start() {
        message = ""
        recording = true
        monitor = NSEvent.addLocalMonitorForEvents(matching: .keyDown) { e in
            record(e)
            return nil   // the keys go to the recorder, not to the window
        }
    }

    private func stop() {
        if let m = monitor { NSEvent.removeMonitor(m) }
        monitor = nil
        recording = false
    }

    private func record(_ e: NSEvent) {
        let f = e.modifierFlags.intersection(.deviceIndependentFlagsMask)
        if e.keyCode == 53 && f.isEmpty { stop(); return }                          // Esc
        let s = Shortcut(keyCode: Int(e.keyCode), key: Self.keyName(e),
                         control: f.contains(.control), option: f.contains(.option),
                         shift: f.contains(.shift), command: f.contains(.command))
        guard s.isUsable else {
            message = f.contains(.command) ? "⌘ can't be used" : "Hold Control or Option"
            return
        }
        if let other = SharedSettings.action(using: s, except: action) {
            message = "Already used for \(other.title)"
            return
        }
        save(s)
        stop()
    }

    private func save(_ s: Shortcut?) {
        SharedSettings.setShortcut(s, for: action)
        current = s
        message = ""
    }

    static func keyName(_ e: NSEvent) -> String {
        let names: [UInt16: String] = [49: "Space", 36: "↩", 48: "⇥", 51: "⌫", 117: "⌦", 123: "←", 124: "→",
                                       125: "↓", 126: "↑", 115: "↖", 119: "↘", 116: "⇞", 121: "⇟",
                                       122: "F1", 120: "F2", 99: "F3", 118: "F4", 96: "F5", 97: "F6",
                                       98: "F7", 100: "F8", 101: "F9", 109: "F10", 103: "F11", 111: "F12"]
        if let n = names[e.keyCode] { return n }
        return (e.charactersIgnoringModifiers ?? "?").uppercased()
    }
}
