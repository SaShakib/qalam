// The live typing pad: the same engine as the apps (Go, compiled to WebAssembly), with the
// options panel, ↑↓ and Space, like the real keyboard.
(() => {
  const pad = document.getElementById("pad");
  if (!pad) return;
  const textEl = pad.querySelector(".text");
  const latinEl = document.getElementById("pad-latin");
  const optsEl = document.getElementById("options");
  const hint = pad.querySelector(".hint");
  const sel = (id) => document.getElementById(id);
  const touch = matchMedia("(pointer: coarse)").matches && !matchMedia("(pointer: fine)").matches;

  let committed = "", buffer = "", cands = [], pick = 0, ready = false;

  const opts = () => ({
    style: sel("o-style").value,
    harakat: sel("o-harakat").value,
    sukun: sel("o-sukun").value,
    spelling: true,
    digits: false,
  });

  function refresh() {
    cands = buffer ? window.qalam.candidates(buffer, opts()) : [];
    if (pick >= cands.length) pick = 0;
    render();
  }

  function current() {
    const c = cands[pick];
    return c && c.text ? c.text : buffer;
  }

  function render() {
    textEl.textContent = "";
    textEl.append(document.createTextNode(committed));
    if (buffer) {
      const w = document.createElement("span");
      w.className = "word";
      w.textContent = current();
      textEl.append(w);
    }
    const caret = document.createElement("span");
    caret.className = "caret";
    textEl.append(caret);
    latinEl.textContent = buffer ? buffer : "";
    hint.hidden = !!(committed || buffer);
    optsEl.innerHTML = "";
    if (!buffer || !sel("o-panel").checked) {
      optsEl.innerHTML = `<div class="empty">${buffer ? "Options are off." : "Options for the word you type show here. ↑ ↓ choose, Space puts it in."}</div>`;
      return;
    }
    cands.forEach((c, i) => {
      const row = document.createElement("div");
      row.className = "opt" + (i === pick ? " on" : "");
      row.innerHTML = `<span class="n">${i + 1}</span><span class="t" lang="ar"></span><span class="l"></span>`;
      row.querySelector(".t").textContent = c.text;
      row.querySelector(".l").textContent = c.label;
      row.onmousedown = (e) => { e.preventDefault(); pick = i; commit(); render(); pad.focus(); };
      optsEl.append(row);
    });
  }

  function commit() {
    if (!buffer) return;
    committed += current();
    buffer = ""; cands = []; pick = 0;
  }

  // One key, as the keyboard sees it. Returns true when the pad used it.
  function key(k, shift) {
    if (k === "Backspace") {
      if (buffer) { buffer = buffer.slice(0, -1); refresh(); }
      else { committed = [...committed].slice(0, -1).join(""); render(); }
      return true;
    }
    if (k === "Escape") {
      if (!buffer) return false;
      committed += buffer; buffer = ""; cands = []; render();
      return true;
    }
    if (k === "Enter") { commit(); committed += "\n"; render(); return true; }
    if (k === " ") { commit(); committed += " "; render(); return true; }
    if ((k === "ArrowDown" || k === "ArrowUp") && buffer && cands.length) {
      pick = (pick + (k === "ArrowDown" ? 1 : cands.length - 1)) % cands.length;
      render();
      return true;
    }
    if ([...k].length !== 1) return false;
    let ch = k;
    // Case comes from Shift only (as in the apps), so Caps Lock can't turn س into ص.
    if (/^[a-z]$/i.test(ch)) ch = shift ? ch.toUpperCase() : ch.toLowerCase();
    if (window.qalam.accepts(ch, buffer)) {
      buffer += ch;
      refresh();
    } else {
      commit();
      committed += window.qalam.punctuation(ch, opts()) || ch;
      render();
    }
    return true;
  }

  pad.addEventListener("keydown", (e) => {
    if (!ready || e.isComposing) return;
    if (e.ctrlKey || e.metaKey || e.altKey) {
      if (e.key.toLowerCase() === "c" && (e.ctrlKey || e.metaKey)) { copy(); e.preventDefault(); }
      return;
    }
    if (key(e.key, e.shiftKey)) e.preventDefault();
  });
  pad.addEventListener("blur", () => { commit(); render(); });

  function copy() {
    commit(); render();
    navigator.clipboard && navigator.clipboard.writeText(committed).then(() => flash("Copied ✓"), () => flash("Select and copy by hand"));
  }
  function flash(t) { const b = sel("pad-copy"); const old = b.textContent; b.textContent = t; setTimeout(() => b.textContent = old, 1400); }
  sel("pad-copy").onclick = copy;
  sel("pad-clear").onclick = () => { committed = ""; buffer = ""; cands = []; render(); pad.focus(); };
  for (const id of ["o-style", "o-harakat", "o-sukun", "o-panel"]) sel(id).addEventListener("change", () => { refresh(); pad.focus(); });

  // Sample chips type a word for you, one key at a time.
  let typing = false;
  document.querySelectorAll("[data-type]").forEach((chip) => {
    chip.onclick = async () => {
      if (!ready || typing) return;
      typing = true;
      pad.focus();
      for (const k of chip.dataset.type) {
        key(k, k !== k.toLowerCase());
        await new Promise((r) => setTimeout(r, 110));
      }
      await new Promise((r) => setTimeout(r, 450));
      key(" ");
      typing = false;
    };
  });

  // Phones and tablets: their keyboards don't send keys one by one, so convert what's typed instead.
  function mobile() {
    const box = document.getElementById("mobile-demo");
    if (!box) return;
    box.hidden = false;
    const input = box.querySelector("textarea"), out = box.querySelector(".out");
    const update = () => { out.textContent = window.qalam.text(input.value, opts()); };
    input.addEventListener("input", update);
    for (const id of ["o-style", "o-harakat", "o-sukun"]) sel(id).addEventListener("change", update);
    update();
  }

  function start() {
    ready = true;
    document.getElementById("demo-loading").hidden = true;
    if (touch) { pad.hidden = true; optsEl.hidden = true; mobile(); }
    render();
  }

  window.addEventListener("qalam-ready", start, { once: true });
  if (!window.Go || !WebAssembly) {
    document.getElementById("demo-loading").textContent = "This browser can't run the demo.";
    return;
  }
  const go = new Go();
  const load = WebAssembly.instantiateStreaming
    ? WebAssembly.instantiateStreaming(fetch("demo/qalam.wasm"), go.importObject)
    : fetch("demo/qalam.wasm").then((r) => r.arrayBuffer()).then((b) => WebAssembly.instantiate(b, go.importObject));
  load.then((r) => go.run(r.instance)).catch(() => {
    document.getElementById("demo-loading").textContent = "The demo couldn't load. Refresh the page to try again.";
  });
})();
