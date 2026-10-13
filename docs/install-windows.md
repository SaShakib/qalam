# Install on Windows

Needs **Windows 10 or 11**. No administrator rights are needed: Qalam installs for your user only.

## 1. Download the right file

| Your PC | File |
|---|---|
| Most PCs (Intel or AMD, "x64-based processor") | **`Qalam-Setup-x64.exe`** |
| ARM PCs (Snapdragon, Surface Pro X, "ARM-based processor") | **`Qalam-Setup-arm64.exe`** |

The **Download** button on the [website](https://sashakib.github.io/qalam/) picks the right one. To check by hand: **Settings → System → About → System type**.

## 2. Run the installer

1. Open the downloaded file.
2. If **"Windows protected your PC"** appears (Microsoft SmartScreen), click **More info**, then **Run anyway**. This shows because the installer isn't code-signed yet. It only happens for the first install, not for updates.
3. Choose whether Qalam should **start when you sign in** (recommended) and whether you want a desktop shortcut.
4. Click **Install**, then **Finish**. Qalam starts and opens its [setup page](getting-started.md) in your browser.

The installer puts Qalam in `%LOCALAPPDATA%\Programs\Qalam`, adds it to the Start menu, and lists it in **Settings → Apps** so you can uninstall it like any app.

## 3. Find Qalam in the tray

Qalam lives in the **system tray**, bottom right of the taskbar (click **^** if you don't see it). The **ق** icon is green when Arabic typing is on.

- **Ctrl + Alt + A** switches Arabic on and off (or click the ق icon).
- **Right-click ق** for every setting, including **Setup & settings…**

> Drag the ق icon from the **^** area onto the taskbar so it's always visible.

## No-installer version

If you can't run installers, download **`Qalam-Windows-x64.exe`** (or `-arm64.exe`) instead. It's the whole app in one file:

- Run it. It asks **"Install Qalam for this user?"** Choose **Yes** to copy it to your Programs folder, the Start menu and startup, or **No** to just run it from where it is.
- It updates itself the same way.

## Silent install (for IT and testing)

```
Qalam-Setup-x64.exe /VERYSILENT /SUPPRESSMSGBOXES /NORESTART
```

Add `/TASKS="startup"` or `/TASKS=""` to choose whether it starts with Windows. To uninstall silently, run `%LOCALAPPDATA%\Programs\Qalam\unins000.exe /VERYSILENT`.

## Things to know

- **Apps running as administrator** (for example, a Command Prompt opened with "Run as administrator") don't accept typing from normal apps, so Qalam can't type there. This is a Windows security rule.
- **Chrome, Edge, and apps built on them** (VS Code, Slack, Discord, WhatsApp, Teams…) get a small automatic fix so اللَّه keeps its harakat.
- Your antivirus might ask about Qalam because it watches the keyboard to type Arabic. That's what every keyboard app does. Qalam never sends anything you type anywhere.

Problems? See [Troubleshooting](troubleshooting.md#windows). To remove Qalam, see [Uninstall](uninstall.md).
