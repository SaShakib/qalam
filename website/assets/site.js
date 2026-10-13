// Shared by every page: header, footer, theme switch, GitHub helpers.
(() => {
  const site = window.QALAM_SITE || { repo: "SaShakib/qalam" };
  const repoUrl = `https://github.com/${site.repo}`;
  const page = document.body.dataset.page || "";

  const links = [
    ["index.html#download", "Download", "download"],
    ["docs.html?p=install", "Install", "install", "hide-sm"],
    ["index.html#demo", "Try it", "demo", "hide-sm"],
    ["docs.html", "Docs", "docs"],
    ["questions.html", "Questions", "questions"],
  ];
  const head = document.createElement("header");
  head.className = "site-head";
  head.innerHTML = `<div class="wrap">
    <a class="brand" href="index.html"><span class="ar" lang="ar">قَلَم</span><span>Qalam</span></a>
    <nav class="nav">${links.map(([href, text, id, cls]) =>
      `<a href="${href}" class="${cls || ""} ${id === page ? "on" : ""}">${text}</a>`).join("")}
      <a href="${repoUrl}" class="hide-sm" rel="noopener">GitHub</a>
      <button class="theme-btn" type="button" title="Light or dark" aria-label="Switch light or dark">◐</button>
    </nav></div>`;
  document.body.prepend(head);

  const foot = document.createElement("footer");
  foot.className = "site-foot";
  foot.innerHTML = `<div class="wrap">
    <div><span class="ar" lang="ar">قَلَم</span> Qalam · free, open source, no tracking. Works offline.</div>
    <div><a href="docs.html">Docs</a> · <a href="questions.html">Questions</a> · <a href="${repoUrl}/releases">All releases</a> ·
    <a href="${repoUrl}/issues/new?template=bug_report.yml">Report a bug</a> · <a href="${repoUrl}">Source</a></div></div>`;
  document.body.append(foot);

  // Theme: follows the system unless the visitor picks one (remembered on this browser only).
  const root = document.documentElement;
  try { const t = localStorage.getItem("qalam-theme"); if (t) root.dataset.theme = t; } catch {}
  head.querySelector(".theme-btn").onclick = () => {
    const dark = root.dataset.theme ? root.dataset.theme === "dark" : matchMedia("(prefers-color-scheme: dark)").matches;
    root.dataset.theme = dark ? "light" : "dark";
    try { localStorage.setItem("qalam-theme", root.dataset.theme); } catch {}
  };

  // GitHub API with a short cache, so moving between pages doesn't use up the hourly limit.
  async function gh(path, maxAgeMs = 5 * 60 * 1000) {
    const key = "gh:" + path;
    try {
      const hit = JSON.parse(sessionStorage.getItem(key) || "null");
      if (hit && Date.now() - hit.t < maxAgeMs) return hit.v;
    } catch {}
    const r = await fetch(`https://api.github.com/repos/${site.repo}${path}`, { headers: { Accept: "application/vnd.github+json" } });
    if (!r.ok) throw new Error(`GitHub answered ${r.status}`);
    const v = await r.json();
    try { sessionStorage.setItem(key, JSON.stringify({ t: Date.now(), v })); } catch {}
    return v;
  }

  function esc(s) {
    return String(s).replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
  }

  window.Qalam = { site, repoUrl, gh, esc };
})();
