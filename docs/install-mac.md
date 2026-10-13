# Install on Mac

Needs **macOS 13 Ventura or later**. One download works on both Apple Silicon (M1 and later) and Intel Macs.

There are three ways to install. Pick **one**.

## Option 1: the installer package (easiest)

1. Download **`Qalam-mac.pkg`** from the [website](https://sashakib.github.io/qalam/) or the [releases page](https://github.com/SaShakib/qalam/releases/latest).
2. Double-click it in your Downloads folder.
3. Click **Continue** and **Install**, and enter your Mac password when asked.
4. When it finishes, the **Qalam** app opens by itself with the [guided setup](getting-started.md).

### If macOS says "Apple could not verify Qalam…"

Until Qalam's installer is notarized by Apple, macOS blocks it the first time. This happens **once**, not for every update:

1. Click **Done** (not "Move to Bin").
2. Open **System Settings → Privacy & Security**.
3. Scroll down to the message *"Qalam-mac.pkg" was blocked…* and click **Open Anyway**.
4. Enter your password, then click **Open Anyway** again. The installer starts.

> On macOS 15 Sequoia and later, Control-clicking the file and choosing **Open** no longer skips this check. Use **System Settings → Privacy & Security** as above.

## Option 2: Homebrew

If you use [Homebrew](https://brew.sh), paste these one at a time in Terminal:

```bash
brew tap sashakib/qalam https://github.com/SaShakib/qalam
```

```bash
brew trust sashakib/qalam
```

```bash
brew install --cask qalam
```

Homebrew 7 asks you to trust a tap before installing from it (the second command). Then open **Qalam** from Applications for the setup. Homebrew installs never show the "could not verify" message.

## Option 3: one Terminal command (no password)

Installs for your user only, into `~/Applications` and `~/Library/Input Methods`:

```bash
curl -fsSL https://raw.githubusercontent.com/SaShakib/qalam/main/install.sh | bash
```

The script downloads the latest release, installs it and opens the setup. This way never shows the "could not verify" message either.

## After installing

The **Qalam** app opens a [guided setup](getting-started.md). Its first step adds the Qalam keyboard to your Mac:

- Click **Add Qalam now**. A ق icon appears in the menu bar's input menu.
- Switch keyboards with the **🌐 Globe** key or **Control-Space**.

If Qalam doesn't show in the input menu, log out and back in once. macOS sometimes needs that for a new keyboard. More fixes are in [Troubleshooting](troubleshooting.md#mac).

## Which one should I choose?

| | Installer package | Homebrew | Terminal command |
|---|---|---|---|
| Installs for | everyone on the Mac | you | you |
| Asks for a password | yes | no | no |
| "Could not verify" warning | once, until notarized | never | never |
| Updates | from the app (asks for your password) | from the app, or `brew upgrade --cask qalam` | from the app, automatic |

To remove Qalam, see [Uninstall](uninstall.md).
