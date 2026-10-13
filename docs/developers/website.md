# Website and questions board

The website is plain HTML, CSS and JavaScript in `website/`, published to **GitHub Pages** by `.github/workflows/pages.yml` on every push to `main` that touches the site, the docs or the engine.

**Address:** https://sashakib.github.io/qalam/

## First-time setup (once)

In the repository on GitHub: **Settings → Pages → Build and deployment → Source: GitHub Actions**. The next push to `main` publishes the site. (A custom domain can be added on the same page.)

## Pages

| Page | File | What |
|---|---|---|
| Home | `website/index.html` | download button (finds the visitor's system and processor), how to install, video, live typing pad, features |
| Docs (the wiki) | `website/docs.html` | renders the Markdown in `docs/` (and `LAYOUT.md`), with search |
| Questions | `website/questions.html` | the questions board |

Settings that change without code are in `website/config.js`:

- **`videoEmbed`**: paste the demo video's embed link (for example `https://www.youtube-nocookie.com/embed/VIDEO_ID`). Empty shows the "coming soon" holder.
- **`questionLabel`**: the issue label the board shows.

## The download button

`website/assets/download.js` reads the system and processor (Chrome and Edge report it exactly; Macs are recognised from the graphics chip), fetches the latest release from the GitHub API and links the matching file:

| Visitor | File |
|---|---|
| Mac | `Qalam-mac.pkg` (universal), else any `.pkg` |
| Windows, ARM64 | `Qalam-Setup-arm64.exe`, else `Qalam-Windows-arm64.exe` |
| Windows, other | `Qalam-Setup-x64.exe`, else `Qalam-Windows-x64.exe` |
| Phone, tablet, Linux | goes to the typing pad |

Windows visitors whose browser can't tell the processor (Firefox) also get an "On an ARM PC?" link. "Other downloads" lists every file in the release.

## The live typing pad

`make wasm` compiles the Go engine (`windows/wasm/`) to `website/demo/qalam.wasm`, so the pad types exactly like the apps. The Pages workflow builds it; it isn't committed.

## The docs and the wiki

All documentation is written **once**, as Markdown files in `docs/`, and shows up in three places:

| Where | Address | How it gets there |
|---|---|---|
| The repository | https://github.com/SaShakib/qalam/tree/main/docs | the files themselves |
| The website | https://sashakib.github.io/qalam/docs.html | `pages.yml` copies `docs/` into the site, `docs.html` renders it with search |
| The GitHub Wiki | https://github.com/SaShakib/qalam/wiki | `wiki.yml` runs `tools/docs_to_wiki.py` on every push to `main` and pushes the pages, with a sidebar |

**Edit `docs/`, never the wiki directly:** the next push to `main` rewrites the wiki's pages. Each wiki page ends with a link to the file it came from.

Docs are Markdown files in `docs/`. To add a page, create `docs/<name>.md` and list it in `docs/nav.json`. Links between pages are ordinary relative Markdown links (`[Fonts](fonts.md)`), so they work on GitHub too. Every page has **Edit this page**, which opens the file on GitHub for a pull request: that's how anyone can improve the wiki.

## The questions board

The board shows **GitHub issues with the `question` label**, as cards. No database or server: the page reads the public GitHub API (cached for 5 minutes per visitor).

- **Ask a question** opens the issue form `.github/ISSUE_TEMPLATE/question.yml`, which adds the `question` label.
- A question shows as **Answered** when its issue is **closed**, or has the **`answered`** label. Create that label once (Issues → Labels → New label) if you want to mark answers but keep the thread open.
- Other labels (for example `mac`, `windows`, `typing`) show on the cards.

**Why GitHub issues?** Free, already where the code is, with notifications, search and moderation (lock, hide, delete). The trade-off is that askers need a GitHub account. If that becomes a barrier, [GitHub Discussions](https://docs.github.com/discussions) with its Q&A category (accepted answers) is the next step up, and [giscus](https://giscus.app) can embed it on the site; a hosted form (for example Formspree into a private inbox) would let people ask without an account but without public answers.

## Preview locally

```bash
make site
python3 -m http.server -d build/site 8000
```

Then open http://localhost:8000.
