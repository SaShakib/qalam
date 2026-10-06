import SwiftUI
import QalamEngine

/// The letter table, the harakat and special keys, then example words.
struct LettersView: View {
    @EnvironmentObject var app: AppState

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                PageHeader(title: "Letters & keys",
                           subtitle: "Spell each word: the letter, then its haraka. Click ▶ to practise a letter.")

                Grid(alignment: .leading, horizontalSpacing: 22, verticalSpacing: 6) {
                    GridRow {
                        Text("Letter"); Text("Name"); Text("Type"); Text("Example"); Text(""); Text("Meaning"); Text("")
                    }
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(.secondary)
                    Divider()
                    ForEach(Content.letters) { row in
                        GridRow {
                            Text(row.arabic)
                                .font(app.arabic(32))
                                .frame(minWidth: 44)
                            Text(row.name)
                            HStack(spacing: 4) { ForEach(row.keys, id: \.self) { KeyCap($0) } }
                            VStack(alignment: .leading, spacing: 2) {
                                Text(row.example).font(.body.monospaced())
                                let also = app.also(row.example)
                                if !also.isEmpty {
                                    Text("or " + also.joined(separator: " · "))
                                        .font(.caption.monospaced()).foregroundStyle(.secondary)
                                }
                            }
                            Text(app.show(row.example)).font(app.arabic(24))
                            Text(row.meaning).foregroundStyle(.secondary)
                            Button { app.practise(.letter(row.arabic)) } label: {
                                Image(systemName: "play.circle.fill").font(.title3)
                            }
                            .buttonStyle(.borderless)
                            .help("Practise \(row.name)")
                        }
                        Divider()
                    }
                }

                SectionTitle("Harakat")
                KeyTable(rows: Content.harakat)

                SectionTitle("Hamza, article and special keys")
                KeyTable(rows: Content.symbols)

                SectionTitle("Example words")
                Text("Type the grey letters to get the Arabic.").foregroundStyle(.secondary)
                WordGrid(items: Content.featured, style: .everyday)
                Button("See all the words →") { app.page = .words }
                    .controlSize(.large)
            }
            .padding(28)
        }
    }
}

struct KeyTable: View {
    @EnvironmentObject var app: AppState
    let rows: [KeyRow]

    var body: some View {
        Grid(alignment: .leading, horizontalSpacing: 22, verticalSpacing: 6) {
            GridRow {
                Text("Shows"); Text("Name"); Text("Type"); Text("Example"); Text("")
            }
            .font(.caption.weight(.semibold))
            .foregroundStyle(.secondary)
            Divider()
            ForEach(rows) { r in
                GridRow {
                    Text(r.shows).font(app.arabic(24)).frame(minWidth: 60, alignment: .leading)
                    Text(r.name)
                    Text(r.keys).font(.body.monospaced())
                    Text(r.example).font(.body.monospaced()).foregroundStyle(.secondary)
                    Text(r.example.isEmpty ? "" : app.show(r.example, style: r.example.contains("=") ? .quran : nil))
                        .font(app.arabic(24))
                }
                Divider()
            }
        }
    }
}

/// Cards: Arabic on top, what to type, meaning.
struct WordGrid: View {
    @EnvironmentObject var app: AppState
    let items: [WordItem]
    let style: Style

    var body: some View {
        LazyVGrid(columns: [GridItem(.adaptive(minimum: 190), spacing: 12)], spacing: 12) {
            ForEach(items) { w in
                VStack(alignment: .leading, spacing: 6) {
                    Text(app.show(w.latin, style: style))
                        .font(app.arabic(26, quran: style == .quran))
                        .frame(maxWidth: .infinity, alignment: .trailing)
                        .lineLimit(2)
                    Text(w.latin).font(.callout.monospaced()).foregroundStyle(.secondary).textSelection(.enabled)
                    let also = app.also(w.latin)
                    if !also.isEmpty {
                        Text("or " + also.joined(separator: " · "))
                            .font(.caption.monospaced()).foregroundStyle(.tertiary)
                    }
                    Text(w.meaning).font(.callout)
                }
                .padding(12)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(RoundedRectangle(cornerRadius: 10).fill(Color.secondary.opacity(0.07)))
            }
        }
    }
}
