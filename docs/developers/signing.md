# Signing and trust

## Why users see warnings

- **Mac:** macOS checks every app downloaded in a browser with **Gatekeeper**. An app that isn't signed with an Apple **Developer ID** and **notarized** by Apple gets "Apple could not verify…", and the user has to go to **Privacy & Security → Open Anyway**. Qalam's builds have been signed *ad hoc* (no identity), so every new version downloaded in a browser looked like a brand-new unknown app and asked again.
- **Windows:** **SmartScreen** warns about downloaded programs that aren't code-signed or haven't built up reputation yet ("Windows protected your PC").

## What Qalam already does (no certificate needed)

| Where | What |
|---|---|
| Mac in-app updater, `.zip` route | clears the quarantine mark from the download and the installed apps, so updates never meet Gatekeeper |
| Mac in-app updater, `.pkg` route | **new:** installs the package with `installer` behind the standard password prompt, instead of opening it in Installer, so there's no "Open Anyway" for updates |
| Homebrew cask | clears the quarantine mark after installing (`postflight`) |
| `install.sh` | downloads with `curl` (no quarantine) and clears it anyway |
| Windows updater | downloads with Go's HTTP client, so the new exe has no "Mark of the Web" and SmartScreen doesn't ask |
| Windows installer | per-user install, no administrator prompt (UAC) |

So users see **one** warning: the first file they download in a browser. The steps for it are in [Install on Mac](../install-mac.md) and [Install on Windows](../install-windows.md).

## Removing the warning completely: Mac

You need an [Apple Developer Program](https://developer.apple.com/programs/) membership (99 USD a year).

1. In your developer account, create two certificates: **Developer ID Application** and **Developer ID Installer**. Install them in your login keychain.
2. Make an [app-specific password](https://support.apple.com/102654) for your Apple ID, then save notarization credentials once:

   ```bash
   xcrun notarytool store-credentials qalam --apple-id you@example.com --team-id ABCDE12345
   ```

3. Release with the identities:

   ```bash
   make release \
     SIGN_ID="Developer ID Application: Your Name (ABCDE12345)" \
     INSTALLER_ID="Developer ID Installer: Your Name (ABCDE12345)" \
     NOTARY_PROFILE=qalam
   ```

   The Makefile then signs both apps with the hardened runtime and a secure timestamp, notarizes and staples the apps and the zip, signs the `.pkg` with the installer identity, and notarizes and staples it.

**In CI:** add these repository secrets (Settings → Secrets and variables → Actions) and the release workflow does the same:

| Secret | Value |
|---|---|
| `MACOS_CERTS_P12` | both certificates with private keys, exported from Keychain Access as one `.p12`, then `base64 -i certs.p12` |
| `MACOS_CERTS_PASSWORD` | the `.p12` password |
| `MACOS_SIGN_ID` | `Developer ID Application: Your Name (ABCDE12345)` |
| `MACOS_INSTALLER_ID` | `Developer ID Installer: Your Name (ABCDE12345)` |
| `APPLE_ID`, `APPLE_TEAM_ID`, `APPLE_APP_PASSWORD` | for `notarytool` |

Once releases are notarized, remove the "Open Anyway" steps from the install docs, and the cask's quarantine step can go too.

## Removing the warning completely: Windows

Sign the exes and installers with an Authenticode certificate. Options, cheapest first:

| Option | Cost | Notes |
|---|---|---|
| [SignPath Foundation](https://signpath.org/) | free for open-source projects | they sign builds made by your GitHub Actions; apply with the repository |
| [Azure Trusted Signing](https://learn.microsoft.com/azure/trusted-signing/) | about 10 USD a month | Microsoft's own; reputation is immediate for verified identities |
| OV or EV code-signing certificate | 200–500 USD a year | needs a hardware token or cloud HSM for new certificates |

**In CI:** with a `.pfx` certificate, add `WINDOWS_CERT_PFX` (`base64` of the `.pfx`) and `WINDOWS_CERT_PASSWORD` secrets. The release workflow signs `Qalam-Windows-*.exe` before building the installers, and then the installers, with `signtool` and a timestamp. For SignPath or Trusted Signing, replace that step with their GitHub Action.

The Windows updater replaces `Qalam.exe` in place, so a signed build stays signed through updates.
