import SwiftUI
import AppKit
import QalamEngine

/// The Qur'an inside Qalam.app (Resources/Quran); in a development run, quran/ in the repo.
enum QuranData {
    static let library: QuranLibrary? = {
        var dirs: [URL] = []
        if let res = Bundle.main.resourceURL { dirs.append(res.appendingPathComponent("Quran")) }
        dirs.append(URL(fileURLWithPath: FileManager.default.currentDirectoryPath).appendingPathComponent("quran"))
        return QuranLibrary(folders: dirs)
    }()
}

/// Read any sūrah with a translation, search the whole Qur'an, copy āyāt, choose what the
/// keyboard's ⌃⇧Q box inserts, and download more translations.
struct QuranView: View {
    @EnvironmentObject var app: AppState
    @State private var surah = 1
    @State private var from = 1
    @State private var to = 7
    @State private var search = ""
    @State private var results: [Passage] = []
    @State private var resultTotal = 0
    @State private var translation = SharedSettings.quranTranslation
    @State private var insert = SharedSettings.quranInsert
    @State private var available: [QuranTranslation] = []
    @State private var downloading: String?
    @State private var downloadError: String?
    @State private var copied: String?

    private var lib: QuranLibrary? { QuranData.library }

    var body: some View {
        if let lib {
            HStack(spacing: 0) {
                surahList(lib).frame(width: 230)
                Divider()
                VStack(alignment: .leading, spacing: 12) {
                    controls(lib)
                    Divider()
                    if search.trimmingCharacters(in: .whitespaces).isEmpty {
                        reader(lib)
                    } else {
                        searchResults(lib)
                    }
                }
                .padding(20)
            }
            .onAppear { available = lib.available() }
        } else {
            Text("The Qur'an text is missing from this copy of Qalam. Reinstall Qalam to get it back.")
                .padding(28)
        }
    }

    // MARK: Left: sūrahs

    @State private var surahFilter = ""

    private func surahList(_ lib: QuranLibrary) -> some View {
        VStack(spacing: 0) {
            TextField("Find a sūrah (baqarah, 36, البقرة)", text: $surahFilter)
                .textFieldStyle(.roundedBorder)
                .padding(10)
            List(selection: Binding(get: { surah }, set: { if let s = $0 { choose(s, lib) } })) {
                ForEach(filteredSurahs(lib)) { s in
                    HStack {
                        Text("\(s.number)").monospacedDigit().foregroundStyle(.secondary).frame(width: 26, alignment: .trailing)
                        VStack(alignment: .leading, spacing: 1) {
                            Text(s.name)
                            Text("\(s.english) · \(s.ayahs)").font(.caption).foregroundStyle(.secondary)
                        }
                        Spacer()
                        Text(s.arabic).font(app.arabic(17, quran: true))
                    }
                    .tag(s.number)
                }
            }
        }
    }

    private func filteredSurahs(_ lib: QuranLibrary) -> [Surah] {
        let f = surahFilter.trimmingCharacters(in: .whitespaces)
        if f.isEmpty { return lib.surahs }
        return lib.parse(query: f).surahs
    }

    private func choose(_ s: Int, _ lib: QuranLibrary) {
        surah = s
        from = 1
        to = lib.surah(s)?.ayahs ?? 1
        search = ""
    }

    // MARK: Top: search, translation, what ⌃⇧Q inserts

    private func controls(_ lib: QuranLibrary) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Image(systemName: "magnifyingglass").foregroundStyle(.secondary)
                TextField("Search the whole Qur'an: الرحمن, ٱلرَّحْمَٰنِ, rahmaan… or go to 2:255", text: $search)
                    .textFieldStyle(.roundedBorder)
                    .onSubmit { goToReference(lib) }
                    .onChange(of: search) { _ in runSearch(lib) }
            }
            HStack(spacing: 14) {
                Picker("Translation", selection: $translation) {
                    Text("None").tag("")
                    ForEach(available) { Text($0.title).tag($0.id) }
                }
                .frame(maxWidth: 380)
                .onChange(of: translation) { SharedSettings.quranTranslation = $0.isEmpty ? "en.sahih" : $0 }
                Menu(downloading.map { "Downloading \($0)…" } ?? "More translations") {
                    ForEach(lib.catalog.filter { t in !available.contains(t) }) { t in
                        Button(t.title) { Task { await download(t, lib) } }
                    }
                }
                .frame(width: 200)
                .disabled(downloading != nil)
                Spacer()
            }
            HStack(spacing: 10) {
                Text("⌃⇧Q inserts:").foregroundStyle(.secondary)
                Picker("", selection: $insert) {
                    ForEach(QuranInsert.allCases) { Text($0.title).tag($0) }
                }
                .pickerStyle(.segmented)
                .frame(width: 400)
                .onChange(of: insert) { SharedSettings.quranInsert = $0 }
                Text("with the translation chosen above").font(.caption).foregroundStyle(.secondary)
            }
            if let downloadError { Text(downloadError).font(.caption).foregroundStyle(.red) }
        }
    }

    // MARK: Reader

    private func reader(_ lib: QuranLibrary) -> some View {
        let s = lib.surah(surah)!
        return VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .firstTextBaseline) {
                Text("\(s.number). \(s.name)").font(.title.bold())
                Text("\(s.english) · \(s.ayahs) āyāt · \(s.type)").foregroundStyle(.secondary)
                Spacer()
                Text(s.arabic).font(app.arabic(30, quran: true))
            }
            HStack(spacing: 8) {
                Text("Āyāt")
                Stepper(value: $from, in: 1...s.ayahs) { Text("\(from)").monospacedDigit() }
                    .onChange(of: from) { if to < $0 { to = $0 } }
                Text("to")
                Stepper(value: $to, in: from...s.ayahs) { Text("\(to)").monospacedDigit() }
                Button("All") { from = 1; to = s.ayahs }
                Spacer()
                copyButton("Copy \(surah):\(from)\(to > from ? "-\(to)" : "")", Passage(surah: surah, from: from, to: to), lib)
                    .buttonStyle(.borderedProminent)
            }
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 14) {
                    ForEach(from...max(from, to), id: \.self) { a in
                        ayahRow(Passage(surah: surah, from: a, to: a), lib)
                        Divider()
                    }
                }
                .padding(.trailing, 8)
            }
        }
    }

    // MARK: Search

    private func runSearch(_ lib: QuranLibrary) {
        let q = search.trimmingCharacters(in: .whitespaces)
        if q.isEmpty || lib.resolve(q) != nil && q.contains(where: \.isNumber) { results = []; resultTotal = 0; return }
        let arabic = q.unicodeScalars.contains { (0x0600...0x06FF).contains($0.value) }
        let ar = arabic ? q : Qalam.text(q, Options(harakat: .none))
        let r = lib.searchArabic(ar, phonetic: !arabic, limit: 300)
        results = r.hits
        resultTotal = r.total
    }

    /// "2:255" or "baqarah 255-257" + Return: open that passage.
    private func goToReference(_ lib: QuranLibrary) {
        guard let p = lib.resolve(search) else { return }
        surah = p.surah
        from = p.from
        to = p.to
        search = ""
    }

    private func searchResults(_ lib: QuranLibrary) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            if let p = lib.resolve(search), search.contains(where: \.isNumber) {
                Button("Open \(lib.surah(p.surah)?.name ?? "") \(p.reference)  (Return)") { goToReference(lib) }
            } else {
                Text(resultTotal == 0 ? "Nothing found. Harakat are optional; Latin letters are read the Qalam way (H = ح, S = ص)."
                     : "\(resultTotal) āyāt" + (resultTotal > results.count ? ", first \(results.count) shown" : ""))
                    .foregroundStyle(.secondary)
            }
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 14) {
                    ForEach(results, id: \.self) { p in
                        ayahRow(p, lib, showSurah: true)
                        Divider()
                    }
                }
                .padding(.trailing, 8)
            }
        }
    }

    // MARK: One āyah

    private func ayahRow(_ p: Passage, _ lib: QuranLibrary, showSurah: Bool = false) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Text(showSurah ? "\(lib.surah(p.surah)?.name ?? "") \(p.reference)" : p.reference)
                    .font(.caption.weight(.semibold)).foregroundStyle(.secondary)
                if showSurah {
                    Button("Open") { surah = p.surah; from = p.from; to = lib.surah(p.surah)?.ayahs ?? p.from; search = "" }
                        .buttonStyle(.link).font(.caption)
                }
                Spacer()
                copyButton("Copy", p, lib).buttonStyle(.borderless).font(.caption)
            }
            Text(shown(lib.arabic(p.surah, p.from) ?? ""))
                .font(app.arabic(28, quran: true))
                .lineSpacing(10)
                .multilineTextAlignment(.trailing)
                .frame(maxWidth: .infinity, alignment: .trailing)
                .textSelection(.enabled)
            if !translation.isEmpty, let t = lib.translation(translation, p.surah, p.from) {
                Text(t)
                    .font(.system(size: 14 * app.fontScale))
                    .foregroundStyle(.secondary)
                    .textSelection(.enabled)
            }
        }
    }

    /// The Qur'an font picked in Settings; KFGQPC Hafs needs three marks swapped (see QuranLibrary.forKFGQPC).
    private func shown(_ s: String) -> String {
        app.quranFontName.uppercased().contains("KFGQPC") ? QuranLibrary.forKFGQPC(s) : s
    }

    // MARK: Copy (plain text, plus rich text in the Qur'an font for Pages, Word, Notes…)

    private func copyButton(_ title: String, _ p: Passage, _ lib: QuranLibrary) -> some View {
        Button(copied == p.reference + title ? "Copied ✓" : title) {
            let mode: QuranInsert = translation.isEmpty ? .arabic : .both
            let text = lib.text(p, mode, translation: translation)
            let rich = NSMutableAttributedString()
            let quranFont = NSFont(name: app.quranFontName, size: 24)
                ?? NSFontManager.shared.font(withFamily: app.quranFontName, traits: [], weight: 5, size: 24)
                ?? NSFont.systemFont(ofSize: 24)
            let rtl = NSMutableParagraphStyle()
            rtl.baseWritingDirection = .rightToLeft
            rtl.alignment = .right
            for (i, block) in text.components(separatedBy: "\n").enumerated() {
                if i > 0 { rich.append(NSAttributedString(string: "\n")) }
                let isArabic = block.unicodeScalars.contains { (0x0600...0x06FF).contains($0.value) }
                rich.append(NSAttributedString(string: isArabic ? shown(block) : block, attributes: isArabic
                    ? [.font: quranFont, .paragraphStyle: rtl]
                    : [.font: NSFont.systemFont(ofSize: 14)]))
            }
            let pb = NSPasteboard.general
            pb.clearContents()
            if let rtf = rich.rtf(from: NSRange(location: 0, length: rich.length)) { pb.setData(rtf, forType: .rtf) }
            pb.setString(text, forType: .string)
            copied = p.reference + title
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) { if copied == p.reference + title { copied = nil } }
        }
    }

    // MARK: Download a translation from tanzil.net

    private func download(_ t: QuranTranslation, _ lib: QuranLibrary) async {
        downloading = t.short
        downloadError = nil
        defer { downloading = nil }
        do {
            let url = URL(string: "https://tanzil.net/trans/?transID=\(t.id)&type=txt-2")!
            let (data, _) = try await URLSession.shared.data(from: url)
            guard let s = String(data: data, encoding: .utf8), QuranLibrary.parse(s) != nil else {
                downloadError = "\(t.short): the download was not a whole Qur'an translation. Try again later."
                return
            }
            let dir = QuranLibrary.userFolder
            try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
            try data.write(to: dir.appendingPathComponent(t.id + ".txt"))
            lib.reload(t.id)
            available = lib.available()
            translation = t.id
        } catch {
            downloadError = "\(t.short): \(error.localizedDescription)"
        }
    }
}
