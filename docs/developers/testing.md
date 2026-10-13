# Testing

## Automatic tests (CI)

Every push and pull request runs `.github/workflows/ci.yml`:

| Job | Runs on | What it checks |
|---|---|---|
| **Engine and Windows app (Linux)** | ubuntu | `go vet`, `go test ./...` (golden cases, options panel parity, setup page security, Chrome-app detection, settings file), the WebAssembly build |
| **Windows x64** and **Windows ARM64** | `windows-latest`, `windows-11-arm` | the same Go tests on real Windows, then builds `Qalam.exe` and runs its self-test, builds `Qalam-Setup.exe`, installs it silently, starts Qalam, checks the files, Start menu and registry, then uninstalls and checks everything is gone |
| **Mac** | `macos-latest` | `make test` (Swift and Go engines), a universal `make bundle`, and checks both apps contain Apple Silicon and Intel code and pass `codesign --verify` |

## The Windows self-test

A built exe can check itself without touching the keyboard or your settings:

```powershell
.\Qalam.exe --selftest result.txt; Get-Content result.txt
```

It writes `ok <version>` or the list of failures: the engine, the Chrome-app fix for اللَّه, the embedded pages, icons and font, and that the settings page refuses requests without its token.

## Manual checklist: Windows

These match the bugs already fixed on the Mac, so both platforms behave the same. Run them on a real PC before a release.

| # | Check | Expected |
|---|---|---|
| 1 | Fresh install with `Qalam-Setup-x64.exe` (SmartScreen: More info → Run anyway) | installs without admin; Qalam starts; the setup page opens in the browser |
| 2 | Go through every setup step, change each setting | the tray menu shows the same choices; typing follows them at once |
| 3 | Setup → Shortcuts → **Change** Arabic on/off, press Ctrl+Alt+Q | the page and the tray menu show Ctrl+Alt+Q; it toggles Arabic |
| 4 | Type `lillaAhi ` in **Chrome**, **Edge** and **VS Code** | لِلَّهِ with one shadda and no extra alif (the ZWJ fix, same as Mac 1.1.3) |
| 5 | Type `lillaAhi ` in **Notepad** and **Word** | لِلَّهِ (no ZWJ in native apps) |
| 6 | Type `qaraA`, press ↓ twice, Space | قَرَأَ; next time `qaraA` starts on that option (remembered) |
| 7 | Caps Lock on, type `sabaHa` | سَبَحَ (Caps Lock doesn't make ص) |
| 8 | Type a word, press Ctrl+C | the word is put in first, then the copy happens |
| 9 | Esc while typing `hello` | the English letters `hello` |
| 10 | Turn Arabic **off**, quit, start Qalam again | still off (the setting is kept) |
| 11 | Turn Arabic **on**, install an update (tray → Check for updates…, or run a newer Setup) | Qalam restarts by itself, Arabic still on, Settings → Apps shows the new version, no SmartScreen prompt |
| 12 | Tray → Uninstall Qalam… | removed from Apps, Start menu and startup; `%APPDATA%\Qalam` gone |
| 13 | Single exe: run `Qalam-Windows-x64.exe`, answer **Yes** | installs to `%LOCALAPPDATA%\Programs\Qalam`, one entry in Apps |
| 14 | ARM64 PC: repeat 1, 4 and 11 with the arm64 files | same results |
| 15 | 150% display scaling | the typing box and its text are sharp and sized right |

## Manual checklist: Mac

| # | Check | Expected |
|---|---|---|
| 1 | Fresh `.pkg` install | the app opens the setup steps |
| 2 | Setup step 2: **Add Qalam now** | ق appears in the input menu and Control-Space finds it |
| 3 | Every setup step | Settings shows the same choices |
| 4 | `lillaAhi ` in Chrome and in Notes | correct in both |
| 5 | Install an update from Settings → Updates (each install method) | no "could not verify" prompt; the keyboard stays on |
| 6 | Intel Mac | the same build runs |
