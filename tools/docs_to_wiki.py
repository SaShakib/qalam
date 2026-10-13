#!/usr/bin/env python3
"""Turn docs/ into GitHub Wiki pages.

    python3 tools/docs_to_wiki.py OUT_DIR

docs/ stays the one place to write documentation (the website shows it too). This writes:
  Home.md                    from docs/README.md
  <Page-Title>.md            one per page in docs/nav.json, named after its title
  Key-Chart.md               from LAYOUT.md
  _Sidebar.md, _Footer.md    the wiki's navigation

Links between docs (`fonts.md`, `../LAYOUT.md`, `developers/signing.md#ci`) are rewritten to wiki
page names; links to other repository files point at GitHub. The .github/workflows/wiki.yml
workflow runs this on every push to main and pushes the result to the wiki.
"""
import json
import posixpath
import re
import sys
import unicodedata
from pathlib import Path

REPO = "SaShakib/qalam"
SITE = "https://sashakib.github.io/qalam/"
ROOT = Path(__file__).resolve().parent.parent
DOCS = ROOT / "docs"


def page_name(title: str) -> str:
    """GitHub Wiki page name for a title: "Install on Mac" → "Install-on-Mac"."""
    ascii_title = unicodedata.normalize("NFKD", title).encode("ascii", "ignore").decode()
    name = re.sub(r"[^\w\s-]", "", ascii_title).strip()
    return re.sub(r"\s+", "-", name)


def load_pages():
    """docs name (as in nav.json, e.g. "developers/signing") → (wiki page name, title, source file)."""
    nav = json.loads((DOCS / "nav.json").read_text())
    pages = {}
    for group in nav:
        for p in group["pages"]:
            name, title = p["name"], p["title"]
            if name == "index":
                pages[name] = ("Home", title, DOCS / "README.md")
            elif name == "LAYOUT":
                pages[name] = ("Key-Chart", title, ROOT / "LAYOUT.md")
            else:
                pages[name] = (page_name(title), title, DOCS / f"{name}.md")
    return nav, pages


LINK = re.compile(r"(?<!!)\[([^\]]*)\]\(([^)\s]+)\)")


def rewrite_links(text: str, doc_name: str, pages) -> str:
    """Rewrite relative links in the page `doc_name` (e.g. "developers/signing")."""
    base = "" if doc_name == "LAYOUT" else posixpath.dirname(doc_name)
    repo_dir = "" if doc_name == "LAYOUT" else posixpath.join("docs", base)

    def fix(m):
        label, href = m.group(1), m.group(2)
        if re.match(r"^[a-z]+:|^#|^/", href, re.I):
            return m.group(0)
        path, _, anchor = href.partition("#")
        anchor = f"#{anchor}" if anchor else ""
        if doc_name == "LAYOUT":
            resolved = posixpath.normpath(path)  # relative to the repository root
            doc = resolved[len("docs/"):-3] if resolved.startswith("docs/") and resolved.endswith(".md") else None
        else:
            resolved = posixpath.normpath(posixpath.join(base, path))
            doc = None
            if path.endswith(".md"):
                if resolved == "../LAYOUT.md":
                    doc = "LAYOUT"
                elif resolved == "README.md":
                    doc = "index"
                elif not resolved.startswith(".."):
                    doc = resolved[:-3]
        if doc in pages:
            return f"[{label}]({pages[doc][0]}{anchor})"
        # Any other repository file: link to it on GitHub.
        repo_path = posixpath.normpath(posixpath.join(repo_dir, path)) if repo_dir else posixpath.normpath(path)
        kind = "tree" if not posixpath.splitext(repo_path)[1] else "blob"
        return f"[{label}](https://github.com/{REPO}/{kind}/main/{repo_path}{anchor})"

    # Leave code (blocks and `inline`) alone.
    parts = re.split(r"(```.*?```|`[^`\n]*`)", text, flags=re.S)
    return "".join(p if p.startswith("`") else LINK.sub(fix, p) for p in parts)


def main(out_dir: str):
    out = Path(out_dir)
    out.mkdir(parents=True, exist_ok=True)
    nav, pages = load_pages()

    for doc_name, (wiki_name, title, src) in pages.items():
        text = rewrite_links(src.read_text(), doc_name, pages)
        source = src.relative_to(ROOT).as_posix()
        note = (f"\n\n---\n*This page is made from [`{source}`](https://github.com/{REPO}/blob/main/{source}). "
                f"Edit that file (not this wiki page): the wiki is updated from it on every push to `main`.*\n")
        (out / f"{wiki_name}.md").write_text(text.rstrip() + note)

    sidebar = [f"**[Qalam website]({SITE})** · [Download]({SITE}#download)\n"]
    for group in nav:
        sidebar.append(f"\n**{group['section']}**\n")
        for p in group["pages"]:
            sidebar.append(f"- [{p['title']}]({pages[p['name']][0]})")
    sidebar.append(f"\n**Ask**\n\n- [Questions board]({SITE}questions.html)\n"
                   f"- [Ask a question](https://github.com/{REPO}/issues/new?template=question.yml)\n"
                   f"- [Report a bug](https://github.com/{REPO}/issues/new?template=bug_report.yml)")
    (out / "_Sidebar.md").write_text("\n".join(sidebar) + "\n")
    (out / "_Footer.md").write_text(
        f"[Website]({SITE}) · [Docs on the website]({SITE}docs.html) · [Questions]({SITE}questions.html) · "
        f"[Releases](https://github.com/{REPO}/releases) · [Source](https://github.com/{REPO})\n")
    print(f"wrote {len(pages)} pages to {out}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit(__doc__)
    main(sys.argv[1])
