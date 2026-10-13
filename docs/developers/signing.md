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

Qalam uses **[SignPath Foundation](https://signpath.org/)**: free code signing for open-source projects. GitHub Actions builds the files, SignPath signs them with a certificate issued to SignPath Foundation, and the project owner approves each release. The policy is in [Code signing and privacy](../code-signing-policy.md).

### 1. Apply (once)

1. Go to https://signpath.org/apply and apply with the repository https://github.com/SaShakib/qalam. Qalam meets the conditions: an OSI licence (GPL-3.0, in `LICENSE`), builds made only by GitHub Actions, a published [code signing policy](../code-signing-policy.md) with roles, and a privacy statement.
2. Wait for approval (days to a few weeks). SignPath then creates an organization for you and installs their GitHub app.

### 2. Set up the project in SignPath (once)

Use these exact names: the release workflow refers to them.

| In SignPath | Value |
|---|---|
| Project slug | `qalam` |
| Signing policy slug | `release-signing` (the one using the SignPath Foundation certificate; approval required) |
| Artifact configuration `exes` | a zip file containing `Qalam-Windows-x64.exe` and `Qalam-Windows-arm64.exe`, each **Authenticode-signed** |
| Artifact configuration `installers` | a zip file containing `Qalam-Setup-x64.exe` and `Qalam-Setup-arm64.exe`, each **Authenticode-signed** |
| Trusted build system | GitHub.com, for the repository `SaShakib/qalam` |

An artifact configuration for this looks like:

```xml
<artifact-configuration xmlns="http://signpath.io/artifact-configuration/v1">
  <zip-file>
    <pe-file path="*.exe" max-matches="unbounded">
      <authenticode-sign />
    </pe-file>
  </zip-file>
</artifact-configuration>
```

### 3. Connect GitHub (once)

In the repository: **Settings → Secrets and variables → Actions**

- **Variables** tab → `SIGNPATH_ORGANIZATION_ID`: your organization id from SignPath.
- **Secrets** tab → `SIGNPATH_API_TOKEN`: an API token of a SignPath user with submitter rights on the project.

### 4. Every release

Run the release as usual. The Windows job uploads the exes to SignPath and waits. **Approve the signing request in SignPath** (you get an email), and the signed files come back into the release; the same happens for the installers. Without the variable, releases still build, unsigned.

### Other certificates

With your own `.pfx` certificate instead (for example Certum's open-source certificate), add the secrets `WINDOWS_CERT_PFX` (`base64` of the `.pfx`) and `WINDOWS_CERT_PASSWORD`: the release workflow then signs with `signtool`.

Signing shows Qalam's publisher instead of "Unknown publisher". SmartScreen's "Windows protected your PC" can still appear for the first weeks after a release, until enough people have installed it; Microsoft builds reputation from downloads.

The Windows updater replaces `Qalam.exe` in place, so a signed build stays signed through updates.
