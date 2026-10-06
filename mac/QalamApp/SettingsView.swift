import SwiftUI
import QalamEngine

struct SettingsView: View {
    @EnvironmentObject var app: AppState
    @State private var enabled = InputSource.isEnabled

    var body: some View {
        Form {
            Section("Keyboard: applies in every app") {
                Picker("Writing style", selection: $app.options.style) {
                    ForEach(Style.allCases) { Text($0.title).tag($0) }
                }
                Picker("Harakat", selection: $app.options.harakat) {
                    ForEach(HarakatMode.allCases) { Text($0.title).tag($0) }
                }
                Toggle("Smart spelling words (اللَّه، هَٰذَا، ذَٰلِكَ، الَّذِي …)", isOn: $app.options.spellingWords)
                Toggle("Arabic digits (١٢٣)", isOn: $app.options.arabicDigits)
                Toggle("Qur'an style: Madinah-mushaf sukūn (قۡ instead of قْ)", isOn: $app.options.quranSmallSukun)
                Text(app.show("alHamdu lillaAhi rabbi alea^lamiyna"))
                    .font(app.arabic(26))
                    .frame(maxWidth: .infinity, alignment: .trailing)
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
