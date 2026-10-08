import Foundation
import QalamEngine

/// Builds a single self-contained HTML page: starter practice (required on first visit),
/// practice sets, the letter table and the word list. All Arabic is produced here by the engine.
enum Guide {
    struct Item: Encodable {
        let latin: String
        let ar: String
        let meaning: String
        let label: String
        let also: [String]
        let accept: [String]
        let quran: Bool
    }

    struct Letter: Encodable {
        let ar, name, example, exampleAr, meaning: String
        let keys, also: [String]
    }

    struct KeyInfo: Encodable { let shows, name, keys, example, exampleAr: String }

    struct Group: Encodable {
        let id, title, note: String
        let items: [Item]
    }

    struct Data: Encodable {
        let platform: String
        let starter: [Item]
        let letters: [Letter]
        let letterSets: [String: [Item]]
        let harakat: [KeyInfo]
        let symbols: [KeyInfo]
        let groups: [Group]
    }

    static func show(_ latin: String, _ style: Style) -> String {
        var o = Options()
        o.style = style
        return Qalam.text(latin, o)
    }

    static func accept(_ latin: String) -> [String] {
        var out = Set<String>()
        for style in Style.allCases {
            for harakat in HarakatMode.allCases {
                for small in [false, true] {
                    let o = Options(style: style, harakat: harakat, quranSmallSukun: small)
                    out.insert(Qalam.text(latin, o).precomposedStringWithCanonicalMapping)
                }
            }
        }
        return out.sorted()
    }

    static func item(_ p: PracticeItem) -> Item {
        Item(latin: p.latin, ar: show(p.latin, p.style), meaning: p.meaning, label: p.label,
             also: Content.alsoTypes(p.latin), accept: accept(p.latin), quran: p.style == .quran)
    }

    static func keyInfo(_ k: KeyRow) -> KeyInfo {
        KeyInfo(shows: k.shows, name: k.name, keys: k.keys, example: k.example,
                exampleAr: k.example.isEmpty ? "" : show(k.example, k.example.contains("=") ? .quran : .everyday))
    }

    static func fontFace(_ family: String, _ file: String) -> String {
        guard let d = FileManager.default.contents(atPath: file) else { return "" }
        return "@font-face{font-family:'\(family)';src:url(data:font/ttf;base64,\(d.base64EncodedString())) format('truetype');}\n"
    }

    static func build(platform: String, fontsDir: String) -> String {
        var sets: [String: [Item]] = [:]
        for row in Content.letters { sets[row.arabic] = Content.letterDrills(for: row.arabic).map(item) }
        let data = Data(
            platform: platform,
            starter: Content.starter.map(item),
            letters: Content.letters.map {
                Letter(ar: $0.arabic, name: $0.name, example: $0.example, exampleAr: show($0.example, .everyday),
                       meaning: $0.meaning, keys: $0.keys, also: Content.alsoTypes($0.example))
            },
            letterSets: sets,
            harakat: Content.harakat.map(keyInfo),
            symbols: Content.symbols.map(keyInfo),
            groups: Content.groups.map { g in
                Group(id: g.id, title: g.title, note: g.note,
                      items: Content.items(ofGroup: g.id).map(item))
            }
        )
        let enc = JSONEncoder()
        enc.outputFormatting = [.withoutEscapingSlashes]
        let json = String(data: try! enc.encode(data), encoding: .utf8)!
            .replacingOccurrences(of: "</", with: "<\\/")
        let fonts = fontFace("Noto Naskh Arabic", fontsDir + "/NotoNaskhArabic.ttf")
            + fontFace("Amiri Quran", fontsDir + "/AmiriQuran.ttf")
        return page.replacingOccurrences(of: "/*FONTS*/", with: fonts)
            .replacingOccurrences(of: "/*DATA*/", with: json)
    }

    static let page = #"""
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Qalam Guide</title>
<style>
/*FONTS*/
:root{--bg:#fbfaf7;--card:#ffffff;--ink:#1d1d1f;--muted:#6b6b70;--line:#e6e3dc;--accent:#0b6b66;--accent-soft:#e3f1ef;--warn:#b5651d;--good:#1f8a4c;--bad:#c0392b}
@media (prefers-color-scheme: dark){:root{--bg:#161616;--card:#202020;--ink:#f2f2f2;--muted:#a0a0a5;--line:#333;--accent:#4fc1b6;--accent-soft:#1f3533;--warn:#e0a060;--good:#4cc27a;--bad:#e06a5c}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--ink);font:16px/1.5 system-ui,-apple-system,"Segoe UI",sans-serif}
.ar{font-family:"Noto Naskh Arabic","Segoe UI",sans-serif;direction:rtl}
.qu{font-family:"Amiri Quran","Noto Naskh Arabic",serif;direction:rtl}
header{display:flex;align-items:center;gap:16px;padding:14px 24px;border-bottom:1px solid var(--line);background:var(--card);position:sticky;top:0;z-index:2}
header .logo{font-size:40px;color:var(--accent);line-height:1}
header h1{font-size:20px;margin:0}
header p{margin:0;color:var(--muted);font-size:14px}
nav{display:flex;gap:6px;margin-left:auto;flex-wrap:wrap}
nav button{border:1px solid var(--line);background:transparent;color:var(--ink);padding:6px 12px;border-radius:8px;cursor:pointer;font:inherit;font-size:14px}
nav button.on{background:var(--accent);color:#fff;border-color:var(--accent)}
main{max-width:1100px;margin:0 auto;padding:20px 24px 60px}
.banner{background:var(--accent-soft);border-radius:12px;padding:14px 18px;margin-bottom:18px}
.card{background:var(--card);border:1px solid var(--line);border-radius:14px;padding:22px;text-align:center}
.target{font-size:72px;line-height:1.5;margin:4px 0}
.meaning{color:var(--muted);font-size:18px}
.hint{display:inline-block;background:var(--accent-soft);border-radius:10px;padding:10px 18px;margin:14px 0}
.hint code{font-size:26px;font-weight:600}
code,kbd{font-family:ui-monospace,Consolas,monospace}
kbd{border:1px solid var(--line);border-bottom-width:2px;border-radius:5px;padding:1px 6px;background:var(--card);font-size:14px}
.tip{color:var(--warn);font-size:15px;margin-top:6px}
input.answer{width:min(560px,100%);font-size:30px;padding:10px 14px;border:2px solid var(--line);border-radius:10px;background:var(--bg);color:var(--ink);text-align:center}
input.answer:focus{outline:none;border-color:var(--accent)}
.ok{color:var(--good);font-weight:700;font-size:22px}
.no{color:var(--bad);font-weight:700}
.row{display:flex;gap:10px;justify-content:center;flex-wrap:wrap;margin-top:12px}
button.act{border:1px solid var(--line);background:var(--card);color:var(--ink);padding:8px 16px;border-radius:9px;cursor:pointer;font:inherit}
button.primary{background:var(--accent);color:#fff;border-color:var(--accent)}
progress{width:100%;height:10px;accent-color:var(--accent)}
table{width:100%;border-collapse:collapse;background:var(--card);border-radius:12px;overflow:hidden}
th,td{padding:8px 12px;border-bottom:1px solid var(--line);text-align:left;vertical-align:middle}
th{font-size:12px;text-transform:uppercase;letter-spacing:.04em;color:var(--muted)}
td.big{font-size:30px}
td.ex{font-size:24px}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(200px,1fr));gap:12px}
.w{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:12px}
.w .a{font-size:28px;text-align:right}
.w code{color:var(--muted)}
.small{font-size:13px;color:var(--muted)}
h2{margin:28px 0 10px}
select{font:inherit;padding:6px 10px;border-radius:8px;border:1px solid var(--line);background:var(--card);color:var(--ink)}
.hidden{display:none}
</style>
</head>
<body>
<header>
  <div class="logo ar">قَلَم</div>
  <div><h1>Qalam: type Arabic the way you spell it</h1><p id="howto"></p></div>
  <nav id="nav"></nav>
</header>
<main id="main"></main>
<script>
const D = /*DATA*/;
const $ = (s, el=document) => el.querySelector(s);
const esc = s => String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const store = {get(k){try{return localStorage.getItem(k)}catch(e){return null}}, set(k,v){try{localStorage.setItem(k,v)}catch(e){}}};
const toggleKeys = D.platform === 'windows' ? '<kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>A</kbd> (or click the ق tray icon)' : '🌐 or <kbd>Control</kbd> + <kbd>Space</kbd>';
$('#howto').innerHTML = 'Switch Arabic on/off with ' + toggleKeys + '. Type a word, then press <kbd>Space</kbd>.';

let page = 'practice';
let starterDone = store.get('qalamStarterDone') === '1';

function nav(){
  const pages = starterDone ? [['practice','Practice'],['letters','Letters & keys'],['words','Word list'],['help','Help']] : [['practice','Starter practice']];
  $('#nav').innerHTML = pages.map(([id,t]) => `<button class="${id===page?'on':''}" data-p="${id}">${t}</button>`).join('');
  $('#nav').querySelectorAll('button').forEach(b => b.onclick = () => { page = b.dataset.p; render(); });
}

// ---------- practice ----------
let set = 'starter', items = D.starter, idx = 0, solved = false;
function setsList(){
  const s = [['starter','Starter: 20 key words']];
  D.letters.forEach(l => s.push(['L:'+l.ar, 'Letter ' + l.ar + '  ' + l.name]));
  D.groups.forEach(g => s.push(['G:'+g.id, 'Words: ' + g.title]));
  return s;
}
function chooseSet(id){
  set = id; idx = 0; solved = false;
  items = id === 'starter' ? D.starter : id.startsWith('L:') ? D.letterSets[id.slice(2)] : D.groups.find(g => 'G:'+g.id === id).items;
  render();
}
function practice(){
  const mandatory = !starterDone;
  if (mandatory && idx >= items.length) {
    return `<div class="card"><div style="font-size:56px">✅</div><h2>You're ready!</h2><p class="meaning">You've typed every kind of key Qalam uses.</p><button class="act primary" id="finish">Open the letter table and word list</button></div>`;
  }
  const it = items[idx % items.length];
  const head = mandatory
    ? `<div class="banner"><b>A short practice to get familiar.</b> 20 words, each teaching a key you will need. It's optional. <button class="act" id="skipall">Skip practice</button><br><span class="small">Turn Qalam on first: ${toggleKeys}.</span></div>
       <div style="display:flex;gap:12px;align-items:center;margin-bottom:12px"><b style="white-space:nowrap">Word ${idx+1} of ${items.length}</b><progress value="${idx}" max="${items.length}"></progress></div>`
    : `<div style="display:flex;gap:12px;align-items:center;margin-bottom:12px"><select id="setsel">${setsList().map(([id,t]) => `<option value="${id}" ${id===set?'selected':''}>${esc(t)}</option>`).join('')}</select><span class="small">${(idx % items.length)+1} of ${items.length}</span></div>`;
  return head + `<div class="card">
    <div class="target ${it.quran?'qu':'ar'}">${esc(it.ar)}</div>
    <div class="meaning">${esc(it.meaning)}</div>
    <div class="hint">Type <code>${esc(it.latin)}</code> then Space
      ${it.also.length ? `<div class="small">also works: ${it.also.map(a => `<code>${esc(a)}</code>`).join(' · ')}</div>` : ''}
      ${it.label ? `<div class="tip">💡 ${esc(it.label)}</div>` : ''}</div>
    <div><input class="answer ar" id="answer" autocomplete="off" spellcheck="false" placeholder="Type here with Qalam on"></div>
    <div id="fb" style="min-height:40px;margin-top:10px"></div>
    <div class="row">${mandatory ? '' : '<button class="act" id="prev">Previous</button>'}<button class="act" id="skip">${mandatory ? 'Skip this word →' : 'Skip →'}</button></div>
  </div>`;
}
function wirePractice(){
  const done = () => { starterDone = true; store.set('qalamStarterDone','1'); page = 'letters'; render(); };
  if ($('#finish')) { $('#finish').onclick = done; return; }
  if ($('#skipall')) $('#skipall').onclick = done;
  if ($('#setsel')) $('#setsel').onchange = e => chooseSet(e.target.value);
  if ($('#prev')) $('#prev').onclick = () => { idx = Math.max(0, idx-1); render(); };
  if ($('#skip')) $('#skip').onclick = () => { idx++; render(); };
  const box = $('#answer'), fb = $('#fb'), it = items[idx % items.length];
  box.focus();
  const check = final => {
    const v = box.value.trim().normalize('NFC');
    if (!v) { fb.innerHTML = ''; return; }
    if (/[A-Za-z]/.test(v)) { fb.innerHTML = `<span class="no">These are English letters: Qalam is off.</span> <span class="small">Press ${toggleKeys}, clear the box and type again.</span>`; return; }
    if (it.accept.includes(v)) {
      fb.innerHTML = '<span class="ok">✓ Correct!</span>';
      box.disabled = true;
      setTimeout(() => { idx++; render(); }, 700);
    } else if (final || v.endsWith(' ') || box.value.endsWith(' ')) {
      fb.innerHTML = `<span class="no">Not quite.</span> You wrote <span class="ar" style="font-size:24px">${esc(v)}</span>, expected <span class="${it.quran?'qu':'ar'}" style="font-size:24px">${esc(it.ar)}</span>. <button class="act" id="clr">Clear</button>`;
      $('#clr').onclick = () => { box.value=''; fb.innerHTML=''; box.focus(); };
    } else fb.innerHTML = '';
  };
  box.addEventListener('input', () => check(false));
  box.addEventListener('keydown', e => { if (e.key === 'Enter') check(true); });
}

// ---------- reference pages ----------
function letters(){
  const keyRows = rows => `<table><tr><th>Shows</th><th>Name</th><th>Type</th><th>Example</th><th></th></tr>${rows.map(r =>
    `<tr><td class="ex ar">${esc(r.shows)}</td><td>${esc(r.name)}</td><td><code>${esc(r.keys)}</code></td><td><code>${esc(r.example)}</code></td><td class="ex ar">${esc(r.exampleAr)}</td></tr>`).join('')}</table>`;
  return `<h2 style="margin-top:0">Letters & keys</h2><p class="small">Spell each word: the letter, then its haraka. Sukūn and shadda appear by themselves. Click a letter to practise it.</p>
  <table><tr><th>Letter</th><th>Name</th><th>Type</th><th>Example</th><th></th><th>Meaning</th></tr>${D.letters.map(l =>
    `<tr style="cursor:pointer" data-l="${esc(l.ar)}"><td class="big ar">${esc(l.ar)}</td><td>${esc(l.name)}</td><td>${l.keys.map(k => `<kbd>${esc(k)}</kbd>`).join(' ')}</td>
     <td><code>${esc(l.example)}</code>${l.also.length ? `<div class="small">or ${l.also.map(esc).join(' · ')}</div>` : ''}</td><td class="ex ar">${esc(l.exampleAr)}</td><td class="small">${esc(l.meaning)}</td></tr>`).join('')}</table>
  <h2>Harakat</h2>${keyRows(D.harakat)}<h2>Hamza, article and special keys</h2>${keyRows(D.symbols)}`;
}
function words(){
  return D.groups.map(g => `<h2>${esc(g.title)} <button class="act" data-g="${g.id}" style="font-size:14px">Practise these</button></h2><p class="small">${esc(g.note)}</p>
    <div class="grid">${g.items.map(w => `<div class="w"><div class="a ${w.quran?'qu':'ar'}">${esc(w.ar)}</div><code>${esc(w.latin)}</code>${w.also.length ? `<div class="small">or ${w.also.map(esc).join(' · ')}</div>` : ''}<div>${esc(w.meaning)}</div></div>`).join('')}</div>`).join('');
}
function help(){
  return `<h2 style="margin-top:0">How typing works</h2><ul>
  <li>Switch Arabic on/off: ${toggleKeys}.</li>
  <li>While you type a word, it shows in a small preview box. <kbd>Space</kbd>, <kbd>Enter</kbd> or punctuation puts it into your document.</li>
  <li><kbd>Backspace</kbd> removes the last English key; <kbd>Esc</kbd> gives back the English letters.</li>
  <li>Right-click the ق tray icon for Everyday / Qur'an style, harakat, Arabic digits, start with Windows, and uninstall.</li>
  <li>While you type, a panel shows up to 4 versions of the word (as typed, another alif/hamza, no harakat, small alif). <kbd>↑</kbd>/<kbd>↓</kbd> choose, <kbd>Space</kbd> inserts, or click one. Qalam remembers your pick.</li>
  <li>Sukūn is <b>Smart</b>: only at a stop inside a word (مَكْتَب، قُل). Type <code>o</code> to add one yourself. <kbd>Ctrl</kbd>+<kbd>Alt</kbd>+<kbd>O</kbd> (or the panel button) switches sukūn off completely.</li>
  <li>Type a word with no vowels to get bare letters: <code>ktb</code> → كتب.</li>
  <li>Long vowels can be typed two ways: <code>aA</code>=<code>aa</code>, <code>iy</code>=<code>ii</code>, <code>uw</code>=<code>uu</code>. A voweled hamza too: <code>saAala</code>=<code>sa'ala</code>.</li></ul>`;
}

function render(){
  nav();
  const m = $('#main');
  if (!starterDone) page = 'practice';
  m.innerHTML = page === 'practice' ? practice() : page === 'letters' ? letters() : page === 'words' ? words() : help();
  if (page === 'practice') wirePractice();
  m.querySelectorAll('[data-l]').forEach(r => r.onclick = () => { page = 'practice'; chooseSet('L:' + r.dataset.l); });
  m.querySelectorAll('[data-g]').forEach(b => b.onclick = () => { page = 'practice'; chooseSet('G:' + b.dataset.g); });
  window.scrollTo(0, 0);
}
render();
</script>
</body>
</html>
"""#
}
