import SwiftUI
import QalamEngine

struct WordsView: View {
    @EnvironmentObject var app: AppState
    @State private var query = ""

    private func matches(_ w: WordItem) -> Bool {
        let q = query.trimmingCharacters(in: .whitespaces).lowercased()
        return q.isEmpty || w.meaning.lowercased().contains(q) || w.latin.lowercased().contains(q)
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                HStack(alignment: .bottom) {
                    PageHeader(title: "Word list", subtitle: "Every word shows exactly what to type.")
                    Spacer()
                    TextField("Search English or typed form", text: $query)
                        .textFieldStyle(.roundedBorder)
                        .frame(width: 260)
                }
                ForEach(Content.groups) { g in
                    let items = g.items.filter(matches)
                    if !items.isEmpty {
                        HStack(alignment: .firstTextBaseline) {
                            SectionTitle(g.title)
                            Text(g.note).foregroundStyle(.secondary)
                            Spacer()
                            Button("Practise these") { app.practise(.group(g.id)) }
                        }
                        WordGrid(items: items, style: g.style)
                    }
                }
            }
            .padding(28)
        }
    }
}
