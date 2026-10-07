import SwiftUI
import AppKit
import QalamEngine

/// Type with English letters here and copy the Arabic, no keyboard switch needed.
struct WriteView: View {
    @EnvironmentObject var app: AppState
    @AppStorage("writeDraft") private var latin = "dhahaba alwaladu AilaY almadrasat'i."
    @State private var copied = false
    /// Chrome/Electron apps draw الله / لله with doubled marks unless the ligature is blocked.
    @AppStorage("copyForBrowsers") private var copyForBrowsers = true

    private var output: String { Qalam.text(latin, app.options) }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            PageHeader(title: "Write",
                       subtitle: "Type with English letters here. The Arabic appears below, ready to copy. No keyboard switch needed.")
            HStack {
                Picker("Style", selection: $app.options.style) {
                    ForEach(Style.allCases) { Text($0.title).tag($0) }
                }
                .pickerStyle(.segmented)
                .frame(width: 300)
                Spacer()
                Button(copied ? "Copied ✓" : "Copy Arabic") {
                    NSPasteboard.general.clearContents()
                    var o = app.options
                    o.blockAllahLigature = copyForBrowsers
                    NSPasteboard.general.setString(Qalam.text(latin, o), forType: .string)
                    copied = true
                    DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) { copied = false }
                }
                .buttonStyle(.borderedProminent)
                Button("Clear") { latin = "" }
                Toggle("Copy for browsers & chat apps", isOn: $copyForBrowsers)
                    .help("Chrome, the Claude app, VS Code, Slack… draw الله with doubled marks unless this is on. Native Mac apps are fine either way.")
            }
            TextEditor(text: $latin)
                .font(.system(size: 17, design: .monospaced))
                .frame(minHeight: 120, maxHeight: 200)
                .overlay(RoundedRectangle(cornerRadius: 6).stroke(Color.secondary.opacity(0.3)))
            ScrollView {
                Text(output)
                    .font(app.arabic(32))
                    .textSelection(.enabled)
                    .multilineTextAlignment(.trailing)
                    .frame(maxWidth: .infinity, alignment: .trailing)
                    .padding(16)
            }
            .background(RoundedRectangle(cornerRadius: 10).fill(Color.secondary.opacity(0.07)))
        }
        .padding(28)
    }
}
