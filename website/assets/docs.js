// The docs (our wiki): Markdown pages from the repository's docs/ folder, listed in docs/nav.json.
// Anyone can improve a page with "Edit this page" (a pull request on GitHub).
(() => {
  const { repoUrl, esc } = window.Qalam;
  const doc = document.getElementById("doc");
  const toc = document.getElementById("toc");
  let nav = [], pages = [];

  // "install-mac", "developers/signing", "LAYOUT" → docs/<name>.md
  const nameFromHref = (href, from) => {
    if (/^[a-z]+:|^#|^\//i.test(href)) return null;
    const [path, hash] = href.split("#");
    if (!path.endsWith(".md")) return null;
    const base = from.includes("/") ? from.slice(0, from.lastIndexOf("/") + 1) : "";
    const parts = (base + path.replace(/\.md$/, "")).split("/");
    const out = [];
    for (const p of parts) {
      if (p === "..") out.pop(); else if (p && p !== ".") out.push(p);
    }
    let name = out.join("/");
    if (name === "README") name = "index";
    return { name, hash: hash || "" };
  };

  async function fetchPage(name) {
    const r = await fetch(`docs/${name === "index" ? "README" : name}.md`);
    if (!r.ok) throw new Error(r.status);
    return r.text();
  }

  function sourcePath(name) {
    if (name === "LAYOUT") return "LAYOUT.md";
    return `docs/${name === "index" ? "README" : name}.md`;
  }

  async function show(name, hash, push) {
    try {
      const md = await fetchPage(name);
      const html = DOMPurify.sanitize(marked.parse(md));
      doc.innerHTML = `<a class="edit" href="${repoUrl}/edit/main/${sourcePath(name)}" rel="noopener">✎ Edit this page</a>` + html;
    } catch {
      doc.innerHTML = `<h1>Page not found</h1><p>Try the <a href="docs.html">docs home</a> or search on the left.</p>`;
    }
    const h1 = doc.querySelector("h1");
    document.title = (h1 ? h1.textContent : "Docs") + " · Qalam";
    // Headings get anchors; links to other pages stay in the viewer.
    doc.querySelectorAll("h2, h3").forEach((h) => {
      h.id = h.textContent.toLowerCase().trim().replace(/[^\w؀-ۿ]+/g, "-").replace(/^-|-$/g, "");
    });
    doc.querySelectorAll("a[href]").forEach((a) => {
      const t = nameFromHref(a.getAttribute("href"), name);
      if (t) a.href = `docs.html?p=${encodeURIComponent(t.name)}${t.hash ? "#" + t.hash : ""}`;
      else if (/^https?:/.test(a.getAttribute("href"))) a.rel = "noopener";
      else if (!a.getAttribute("href").startsWith("#") && !/^[a-z]+:/i.test(a.getAttribute("href"))) {
        // A file in the repository that isn't a doc page (a script, a cask…): show it on GitHub.
        a.href = `${repoUrl}/blob/main/${(name.includes("/") ? "docs/" + name.slice(0, name.lastIndexOf("/") + 1) : "docs/") + a.getAttribute("href")}`;
      }
    });
    // Arabic in tables and lists reads right to left.
    doc.querySelectorAll("td, li, p").forEach((el) => { if (/[؀-ۿ]/.test(el.textContent)) el.dir = "auto"; });
    toc.querySelectorAll("a").forEach((a) => a.classList.toggle("on", a.dataset.name === name));
    if (push) history.pushState({ name }, "", `docs.html?p=${encodeURIComponent(name)}${hash ? "#" + hash : ""}`);
    if (hash) document.getElementById(hash)?.scrollIntoView(); else window.scrollTo(0, 0);
  }

  document.addEventListener("click", (e) => {
    const a = e.target.closest("a[href^='docs.html?p=']");
    if (!a || e.metaKey || e.ctrlKey || e.shiftKey) return;
    e.preventDefault();
    const u = new URL(a.href);
    show(u.searchParams.get("p"), u.hash.slice(1), true);
    search.value = ""; results.hidden = true;
  });
  window.addEventListener("popstate", () => route(false));

  function route(push) {
    const p = new URLSearchParams(location.search).get("p") || "index";
    show(p, location.hash.slice(1), push);
  }

  // Search: every page is fetched once, then searched in the browser.
  const search = document.getElementById("search"), results = document.getElementById("results");
  let loaded = null;
  search.addEventListener("input", async () => {
    const q = search.value.trim().toLowerCase();
    if (q.length < 2) { results.hidden = true; return; }
    loaded = loaded || Promise.all(pages.map(async (p) => ({ ...p, text: await fetchPage(p.name).catch(() => "") })));
    const all = await loaded;
    const hits = all.filter((p) => p.text.toLowerCase().includes(q) || p.title.toLowerCase().includes(q)).slice(0, 8);
    results.hidden = false;
    results.innerHTML = hits.length ? hits.map((p) => {
      const i = p.text.toLowerCase().indexOf(q);
      const snip = i < 0 ? "" : p.text.slice(Math.max(0, i - 40), i + 60).replace(/[#*`|>\[\]]/g, "");
      return `<a href="docs.html?p=${encodeURIComponent(p.name)}">${esc(p.title)}<small>…${esc(snip)}…</small></a>`;
    }).join("") : `<p class="muted">Nothing found. <a href="questions.html">Ask a question</a>.</p>`;
  });

  fetch("docs/nav.json").then((r) => r.json()).then((n) => {
    nav = n;
    toc.innerHTML = nav.map((g) => `<h4>${esc(g.section)}</h4>` +
      g.pages.map((p) => `<a href="docs.html?p=${encodeURIComponent(p.name)}" data-name="${esc(p.name)}">${esc(p.title)}</a>`).join("")).join("");
    pages = nav.flatMap((g) => g.pages);
    route(false);
  }).catch(() => route(false));
})();
