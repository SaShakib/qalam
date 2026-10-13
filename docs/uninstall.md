# Uninstall

## Mac

First remove Qalam from the input menu: **System Settings → Keyboard → Text Input → Edit…**, select **Qalam**, click **−**.

Then remove the files, depending on how you installed it:

**Homebrew**

```bash
brew uninstall --cask qalam
```

Add `--zap` to remove its settings too.

**Terminal command (`install.sh`)**

```bash
curl -fsSL https://raw.githubusercontent.com/SaShakib/qalam/main/install.sh | bash -s -- --uninstall
```

**Installer package (`.pkg`)**

```bash
sudo rm -rf /Applications/Qalam.app "/Library/Input Methods/QalamInput.app"
```

Settings are in `~/Library/Preferences/com.qalam.shared.plist` and `com.qalam.app.plist`; delete them to start fresh.

## Windows

- **Settings → Apps → Installed apps → Qalam → Uninstall**, or
- right-click **ق** in the tray → **Uninstall Qalam…**

This removes the app, its Start menu entry, starting with Windows, and its settings (`%APPDATA%\Qalam`).

If you only ran the single `.exe` without installing, quit Qalam from the tray and delete the file.
