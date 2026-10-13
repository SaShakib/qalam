# Making a release

## 1. Bump the version

Set `VERSION` in the `Makefile` (one place for every build). Commit.

## 2. Build

**Automatic (recommended).** Push a tag:

```bash
git tag v1.3.0 && git push origin v1.3.0
```

`.github/workflows/release.yml` then:

1. builds the Mac zip and `.pkg` (universal) and the Windows exes on a Mac runner, signed and notarized if the [signing secrets](signing.md) are set;
2. builds `Qalam-Setup-x64.exe` and `Qalam-Setup-arm64.exe` on a Windows runner (Authenticode-signed if a certificate is set) and smoke-tests the installer;
3. creates a **draft** GitHub release with every file and `SHA256SUMS.txt`, and prints the Homebrew cask line in the job summary.

Check the draft, write the notes, and **Publish**. Apps update from the latest published release.

**By hand on a Mac.** `make release` puts the Mac and Windows files in `dist/`. The Windows installers need Windows: run `make windows-installer` there, or use the workflow.

## 3. Update the Homebrew cask

`make release` updates `Casks/qalam.rb` (version and sha256) by itself. With the workflow, copy the `sha256` from the job summary into `Casks/qalam.rb` and commit it after publishing.

## Release files

| File | For |
|---|---|
| `Qalam-mac.pkg` | the website's Mac download (same as the versioned .pkg) |
| `Qalam-<version>.pkg` | Mac installer package |
| `Qalam-mac.zip` | Homebrew, `install.sh` and the in-app updater |
| `Qalam-Setup-x64.exe`, `Qalam-Setup-arm64.exe` | Windows installers (the website's Windows downloads) |
| `Qalam-Windows-x64.exe`, `Qalam-Windows-arm64.exe` | single-file Windows app; the Windows updater downloads these |
| `SHA256SUMS.txt` | checksums |

Keep these names: the updaters and the website look for them.
