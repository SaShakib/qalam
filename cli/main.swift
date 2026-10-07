import Foundation
import QalamEngine

// qalam kataba              → كَتَبَ
// qalam --quran allaAhu     → ٱللَّهُ
// qalam --test tests/cases.tsv
// qalam --words             (every word in the app, for review)

var args = Array(CommandLine.arguments.dropFirst())
var o = Options()

func flags(_ list: [String], into o: inout Options) {
    for f in list {
        switch f {
        case "quran": o.style = .quran
        case "astyped": o.harakat = .asTyped
        case "none": o.harakat = .none
        case "smallsukun": o.quranSmallSukun = true
        case "nospelling": o.spellingWords = false
        case "digits": o.arabicDigits = true
        case "chromium": o.blockAllahLigature = true
        default: break
        }
    }
}

func scalars(_ s: String) -> String {
    s.unicodeScalars.map { String(format: "%04X", $0.value) }.joined(separator: " ")
}

func runTests(_ path: String) -> Int32 {
    guard let text = try? String(contentsOfFile: path, encoding: .utf8) else {
        print("cannot read \(path)"); return 2
    }
    var pass = 0, fail = 0
    for (n, line) in text.components(separatedBy: "\n").enumerated() {
        if line.trimmingCharacters(in: .whitespaces).isEmpty || line.hasPrefix("#") { continue }
        let cols = line.components(separatedBy: "\t")
        guard cols.count >= 2 else { print("line \(n + 1): needs input<TAB>expected"); fail += 1; continue }
        var opt = Options()
        if cols.count >= 3 { flags(cols[2].components(separatedBy: ","), into: &opt) }
        let got = Qalam.text(cols[0], opt)
        if got == cols[1] {
            pass += 1
        } else {
            fail += 1
            print("✗ line \(n + 1): \(cols[0])  [\(cols.count >= 3 ? cols[2] : "everyday")]")
            print("    expected \(cols[1])   \(scalars(cols[1]))")
            print("    got      \(got)   \(scalars(got))")
        }
    }
    print("\(pass) passed, \(fail) failed")
    return fail == 0 ? 0 : 1
}

if args.first == "--test", args.count >= 2 {
    exit(runTests(args[1]))
}

if args.first == "--words" {
    for g in Content.groups {
        var go = Options(); go.style = g.style
        print("\n## \(g.title)")
        for w in g.items { print("\(Qalam.text(w.latin, go))\t\(w.latin)\t\(w.meaning)") }
    }
    print("\n## Letters")
    for l in Content.letters { print("\(l.arabic)\t\(l.keys.joined(separator: " / "))\t\(Qalam.text(l.example, o))\t\(l.example)") }
    print("\n## Starter")
    for it in Content.starter { print("\(Qalam.text(it.latin, o))\t\(it.latin)\t\(it.meaning)\t\(it.label)") }
    print("\n## Letter drills")
    print(Content.letterDrills().map { Qalam.text($0.latin, o) + " " + $0.latin }.joined(separator: "   "))
    exit(0)
}

if args.first == "--guide", args.count >= 2 {
    // qalam --guide out.html [windows|mac]
    let platform = args.count >= 3 ? args[2] : "windows"
    let html = Guide.build(platform: platform, fontsDir: "fonts")
    try! html.write(toFile: args[1], atomically: true, encoding: .utf8)
    print("wrote \(args[1]) (\(html.utf8.count / 1024) KB)")
    exit(0)
}

if args.first == "--also" {
    for w in args.dropFirst() { print(w, "→", Content.alsoTypes(w)) }
    exit(0)
}

if args.first == "--keys", args.count >= 2 {
    // simulate typing, '<' = backspace, '!' = escape, '|' = enter
    var comp = Composer()
    for c in args[1] {
        let key: Composer.Key = c == "<" ? .backspace : c == "!" ? .escape : c == "|" ? .enter : c == " " ? .space : .char(c)
        let (used, acts) = comp.handle(key, o)
        print("\(c): used=\(used) \(acts)")
    }
    exit(0)
}

var words: [String] = []
for a in args {
    if a.hasPrefix("--") { flags([String(a.dropFirst(2))], into: &o) } else { words.append(a) }
}
print(Qalam.text(words.joined(separator: " "), o))
