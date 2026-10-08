import SwiftUI
import QalamEngine

/// Shows an Arabic word and what to type; checks what you typed with the Qalam keyboard.
/// In onboarding mode it runs the 20 starter words once, in order, with no skipping.
struct PracticeView: View {
    @EnvironmentObject var app: AppState
    var onboarding = false
    var onFinish: () -> Void = {}

    @State private var items: [PracticeItem] = []
    @State private var index = 0
    @State private var typed = ""
    @State private var result: Result = .typing
    @State private var showHint = true
    @State private var shuffled = false
    @State private var correct = 0
    @State private var streak = 0
    @State private var best = 0
    @State private var finished = false
    @FocusState private var focused: Bool

    enum Result { case typing, correct, wrong }

    private var item: PracticeItem? { items.isEmpty ? nil : items[index % items.count] }

    private var typedHasLatin: Bool { typed.unicodeScalars.contains { $0.isASCII && CharacterSet.letters.contains($0) } }

    var body: some View {
        VStack(spacing: 0) {
            if onboarding { progressBar } else { toolbar }
            Divider()
            if finished {
                doneCard
            } else if let item {
                ScrollView { card(item).padding(.horizontal, 28).padding(.vertical, 20) }
            } else {
                Spacer()
            }
        }
        .onAppear(perform: load)
        .onChange(of: app.practiceSet) { _ in if !onboarding { load() } }
        .onChange(of: shuffled) { _ in load() }
    }

    private var toolbar: some View {
        HStack(spacing: 16) {
            Picker("Practise", selection: $app.practiceSet) {
                ForEach(app.practiceChoices, id: \.0) { choice in
                    Text(choice.1).tag(choice.0)
                }
            }
            .frame(maxWidth: 340)
            Toggle("Shuffle", isOn: $shuffled)
            Toggle("Show what to type", isOn: $showHint)
            Spacer()
            Label("\(correct)", systemImage: "checkmark.circle").help("Correct answers")
            Label("\(streak)", systemImage: "flame").help("Streak (best \(best))")
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 12)
    }

    private var progressBar: some View {
        HStack(spacing: 14) {
            Text("Word \(min(index + 1, items.count)) of \(items.count)").font(.headline)
            ProgressView(value: Double(index), total: Double(max(items.count, 1)))
            Text("\(index) done").foregroundStyle(.secondary)
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 12)
    }

    private var doneCard: some View {
        VStack(spacing: 18) {
            Spacer()
            Image(systemName: "checkmark.seal.fill").font(.system(size: 64)).foregroundStyle(.green)
            Text("You're ready!").font(.largeTitle.bold())
            Text("You've typed every kind of key Qalam uses. Now type Arabic anywhere: switch with 🌐 or Control-Space.")
                .multilineTextAlignment(.center)
                .foregroundStyle(.secondary)
                .frame(maxWidth: 520)
            Button("Start using Qalam") { onFinish() }
                .buttonStyle(.borderedProminent)
                .controlSize(.large)
            Spacer()
        }
        .frame(maxWidth: .infinity)
    }

    @ViewBuilder
    private func card(_ item: PracticeItem) -> some View {
        VStack(spacing: 16) {
            if !onboarding {
                Text("\(index % items.count + 1) of \(items.count)")
                    .font(.callout).foregroundStyle(.secondary)
            }

            Text(target(item))
                .font(app.arabic(item.latin.count > 14 ? 40 : 68, quran: item.style == .quran))
                .multilineTextAlignment(.center)

            if !item.meaning.isEmpty { Text(item.meaning).font(.title3).foregroundStyle(.secondary) }

            if showHint || result == .wrong || onboarding {
                VStack(spacing: 8) {
                    HStack(spacing: 10) {
                        Text("Type").foregroundStyle(.secondary)
                        Text(item.latin)
                            .font(.system(size: 26, weight: .semibold, design: .monospaced))
                            .textSelection(.enabled)
                        Text("then Space").foregroundStyle(.secondary)
                    }
                    let also = app.also(item.latin)
                    if !also.isEmpty {
                        HStack(spacing: 8) {
                            Text("also works:").foregroundStyle(.secondary)
                            ForEach(also, id: \.self) { KeyCap($0) }
                        }
                    }
                    if !item.label.isEmpty {
                        Label(item.label, systemImage: "lightbulb")
                            .font(.callout)
                            .foregroundStyle(.orange)
                    }
                }
                .padding(.horizontal, 18).padding(.vertical, 10)
                .background(RoundedRectangle(cornerRadius: 10).fill(Color.accentColor.opacity(0.10)))
            } else {
                Button("Show what to type") { showHint = true }
            }

            TextField("Type here with the Qalam keyboard", text: $typed)
                .font(app.arabic(30))
                .multilineTextAlignment(.center)
                .textFieldStyle(.roundedBorder)
                .frame(maxWidth: 560)
                .focused($focused)
                .onSubmit { check(final: true) }
                .onChange(of: typed) { _ in check(final: false) }

            if typedHasLatin {
                VStack(spacing: 4) {
                    Text("You're typing English letters. Qalam reads them as:")
                        .font(.callout).foregroundStyle(.secondary)
                    Text(Qalam.text(typed, latinOptions(item))).font(app.arabic(28))
                    Text("Switch to the Qalam keyboard (🌐 or Control-Space) to type Arabic directly.")
                        .font(.caption).foregroundStyle(.secondary)
                }
            }

            switch result {
            case .correct:
                Label("Correct!", systemImage: "checkmark.circle.fill")
                    .font(.title2.bold()).foregroundStyle(.green)
            case .wrong:
                VStack(spacing: 6) {
                    Label("Not quite. Compare and try again.", systemImage: "xmark.circle.fill")
                        .font(.title3.bold()).foregroundStyle(.red)
                    HStack(spacing: 24) {
                        VStack { Text("you wrote").font(.caption); Text(typedArabic(item)).font(app.arabic(30)) }
                        VStack { Text("expected").font(.caption); Text(target(item)).font(app.arabic(30)) }
                    }
                    Button("Clear and try again") { typed = ""; result = .typing; focused = true }
                }
            case .typing:
                EmptyView()
            }

            HStack {
                if !onboarding { Button("Previous") { move(-1) }.disabled(index == 0) }
                Button("Check") { check(final: true) }
                Button(onboarding ? "Skip this word →" : "Skip →") { move(1) }
            }
            .controlSize(.large)
        }
        .frame(maxWidth: .infinity)
    }

    // MARK: Logic

    private func load() {
        var list = onboarding ? Content.starter : app.items(app.practiceSet)
        if shuffled && !onboarding { list.shuffle() }
        items = list
        index = 0
        typed = ""
        result = .typing
        finished = false
        focused = true
    }

    private func move(_ step: Int) {
        if onboarding && index + step >= items.count {
            index = items.count
            finished = true
            return
        }
        index = max(0, index + step)
        typed = ""
        result = .typing
        focused = true
    }

    private func target(_ item: PracticeItem) -> String {
        app.show(item.latin, style: item.style)
    }

    private func latinOptions(_ item: PracticeItem) -> Options {
        var o = app.options
        o.style = item.style
        o.harakat = .full
        return o
    }

    private func typedArabic(_ item: PracticeItem) -> String {
        let t = typed.trimmingCharacters(in: .whitespacesAndNewlines)
        return typedHasLatin ? Qalam.text(t, latinOptions(item)) : t
    }

    /// Accept the word in either style, with full harakat or with the keyboard's current harakat setting.
    private func acceptable(_ item: PracticeItem) -> [String] {
        var out: [String] = []
        for style in [item.style, item.style == .quran ? Style.everyday : Style.quran] {
            for harakat in [HarakatMode.full, app.options.harakat] {
                for sukun in SukunMode.allCases {
                    for small in [false, true] {
                        var o = app.options
                        o.style = style
                        o.harakat = harakat
                        o.sukun = sukun
                        o.quranSmallSukun = small
                        out.append(Qalam.text(item.latin, o))
                        // any option offered in the panel counts too (single words)
                        if !item.latin.contains(" ") { out += Qalam.candidates(item.latin, o).map(\.text) }
                    }
                }
            }
        }
        return out
    }

    private func check(final: Bool) {
        guard let item, result != .correct else { return }
        let t = typed.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !t.isEmpty else { return }
        let mine = typedArabic(item)
        if acceptable(item).contains(where: { $0 == mine }) {
            result = .correct
            correct += 1
            streak += 1
            best = max(best, streak)
            let current = index
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.8) {
                if index == current, result == .correct { move(1) }
            }
        } else if final {
            result = .wrong
            streak = 0
        } else if result == .wrong {
            result = .typing
        }
    }
}

/// First launch: set up the keyboard, then the 20 starter words. The rest of the app opens after.
struct FirstRunView: View {
    @EnvironmentObject var app: AppState
    @State private var enabled = InputSource.isEnabled
    private let timer = Timer.publish(every: 2, on: .main, in: .common).autoconnect()

    var body: some View {
        VStack(spacing: 0) {
            HStack(alignment: .center, spacing: 18) {
                Text("قَلَم").font(app.arabic(44)).foregroundStyle(Color.accentColor)
                VStack(alignment: .leading, spacing: 4) {
                    Text("Welcome! A short practice to get familiar").font(.title2.bold())
                    Text("20 words, each teaching a key you will need (about 5 minutes). It's optional: skip any time.")
                        .foregroundStyle(.secondary)
                }
                Spacer()
                Button("Skip practice") { app.finishFirstPractice() }
                    .controlSize(.large)
            }
            .padding(.horizontal, 20).padding(.top, 16).padding(.bottom, 10)

            HStack(spacing: 12) {
                if enabled {
                    Label("Qalam keyboard is on. Switch to it with 🌐 or Control-Space (look for ق in the menu bar).",
                          systemImage: "checkmark.circle.fill")
                        .foregroundStyle(.green)
                    Spacer()
                    Button("Switch to Qalam now") { InputSource.select() }
                } else if InputSource.isInstalled {
                    Label("Step 1: turn on the Qalam keyboard.", systemImage: "keyboard")
                        .font(.headline)
                    Spacer()
                    Button("Add Qalam now") { enabled = InputSource.enable(); InputSource.select() }
                        .buttonStyle(.borderedProminent)
                    Button("Keyboard Settings…") { InputSource.openKeyboardSettings() }
                } else {
                    Label("The Qalam keyboard is not installed. You can still practise: type the English letters below.",
                          systemImage: "exclamationmark.triangle.fill")
                        .foregroundStyle(.orange)
                    Spacer()
                }
            }
            .padding(.horizontal, 20).padding(.vertical, 10)
            .background(Color.secondary.opacity(0.07))

            PracticeView(onboarding: true) { app.finishFirstPractice() }
        }
        .onReceive(timer) { _ in if !enabled { enabled = InputSource.isEnabled } }
    }
}
