# Updates

Qalam checks GitHub for a new version when it starts and once a day. It only asks GitHub "what is the latest version?". Nothing about you or your typing is sent.

## Mac

**Settings → Updates** has **Check for updates now**, **Install update**, and switches for automatic checking and installing. A bar at the top of the app tells you when an update is ready.

How the update installs depends on how you installed Qalam:

| You installed with | The update |
|---|---|
| The Terminal command (`install.sh`) | downloads and replaces Qalam in place, by itself if "Install updates automatically" is on |
| Homebrew | runs `brew upgrade --cask qalam` when you click **Install update** |
| The installer package (`.pkg`) | downloads the new package and installs it when you click **Install update**. macOS asks for your password once. |

None of these show the "Apple could not verify…" message: that only appears for a file you download in a browser. If the Qalam keyboard was on before the update, it stays on afterwards.

## Windows

The tray menu has **Check for updates…** (with the version you have) and **Install updates automatically** (on by default).

Updates download the new Qalam, swap it in, and restart Qalam in a few seconds. A notification tells you the new version. SmartScreen doesn't ask again, and your settings stay as they are.

## Check your version

- **Mac:** Qalam → Settings → Updates → **Version**.
- **Windows:** right-click **ق**: it's in **Check for updates… (version …)**.

All versions and what changed in each are on the [releases page](https://github.com/SaShakib/qalam/releases).
