// The questions board: GitHub issues labelled "question", shown as cards.
// Answered = the issue is closed, or has the "answered" label. No database or server needed.
(() => {
  const { site, repoUrl, gh, esc } = window.Qalam;
  const label = site.questionLabel || "question";
  const board = document.getElementById("board");
  const status = document.getElementById("status");
  const q = document.getElementById("q");
  let all = [], filter = "all";

  document.getElementById("ask").href = `${repoUrl}/issues/new?template=question.yml`;

  const answered = (i) => i.state === "closed" || i.labels.some((l) => l.name === "answered");
  const ago = (d) => {
    const s = (Date.now() - new Date(d)) / 1000;
    for (const [n, u] of [[31536000, "year"], [2592000, "month"], [86400, "day"], [3600, "hour"], [60, "minute"]]) {
      if (s >= n) { const k = Math.floor(s / n); return `${k} ${u}${k > 1 ? "s" : ""} ago`; }
    }
    return "just now";
  };

  function render() {
    const text = q.value.trim().toLowerCase();
    const shown = all.filter((i) =>
      (filter === "all" || (filter === "answered") === answered(i)) &&
      (!text || (i.title + " " + (i.body || "")).toLowerCase().includes(text)));
    status.textContent = all.length ? `${shown.length} of ${all.length} questions` : "";
    if (!all.length) {
      status.innerHTML = `No questions yet. <a href="${repoUrl}/issues/new?template=question.yml">Be the first to ask.</a>`;
    }
    board.innerHTML = shown.map((i) => {
      const extra = i.labels.filter((l) => !["question", "answered", label].includes(l.name))
        .map((l) => `<span class="badge label">${esc(l.name)}</span>`).join("");
      return `<a class="q" href="${esc(i.html_url)}" rel="noopener">
        <h3 dir="auto">${esc(i.title.replace(/^\[?question\]?:?\s*/i, ""))}</h3>
        <div class="meta">
          <span class="badge ${answered(i) ? "answered" : "open"}">${answered(i) ? "✓ Answered" : "Waiting for an answer"}</span>
          ${extra}
          <span>💬 ${i.comments}</span>
          <span>${ago(i.updated_at)}</span>
        </div></a>`;
    }).join("");
  }

  document.querySelectorAll(".tabs button").forEach((b) => b.onclick = () => {
    filter = b.dataset.f;
    document.querySelectorAll(".tabs button").forEach((x) => x.classList.toggle("on", x === b));
    render();
  });
  q.addEventListener("input", render);

  gh(`/issues?labels=${encodeURIComponent(label)}&state=all&per_page=100&sort=updated`)
    .then((list) => { all = list.filter((i) => !i.pull_request); render(); })
    .catch(() => {
      status.innerHTML = `<div class="notice">Couldn't load the questions right now (GitHub limits how often a page can ask).
        <a href="${repoUrl}/issues?q=label%3A${encodeURIComponent(label)}">See them on GitHub</a>.</div>`;
    });
})();
