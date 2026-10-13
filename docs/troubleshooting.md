# Troubleshooting

## Mac

### Qalam isn't in the input menu, or Control-Space doesn't find it

1. Open the Qalam app → **Settings** → **Keyboard status**. If it says **In the input menu: No**, click **Add Qalam now**.
2. Or add it by hand: **System Settings → Keyboard → Text Input → Edit… → + → Arabic → Qalam**.
3. Still missing? **Log out and back in** once. macOS sometimes needs that to see a new keyboard.

### Qalam disappeared from the input menu after an update

Open the Qalam app once: it turns the keyboard back on if it was on before. Qalam 1.1 and later do this by themselves.

### "Apple could not verify Qalam is free of malware"

Go to **System Settings → Privacy & Security**, scroll down, and click **Open Anyway**. This happens once for a file downloaded in a browser. See [Install on Mac](install-mac.md#if-macos-says-apple-could-not-verify-qalam).

### Updates keep asking me to allow Qalam

Update from inside the app (**Settings → Updates → Install update**) instead of downloading the new version in a browser: in-app updates don't carry the "downloaded from the internet" mark that makes macOS ask. With the `.pkg`, macOS asks only for your password. Homebrew and the Terminal command never ask. (Once the downloads are notarized by Apple, nobody will be asked at all. See [Signing and trust](developers/signing.md).)

### The underlined word doesn't appear, or letters come out in English

Check the menu bar shows **ق** (Qalam is the current keyboard). Some apps with their own text engines (some games, remote desktops, terminals in raw mode) don't support input methods.

## Windows

### I can't find Qalam

Look for the **ق** icon in the tray, bottom right; click **^** to see hidden icons. If it isn't there, start **Qalam** from the Start menu. If it says "Qalam is already running", it's in the hidden icons.

### Typing comes out in English

Arabic typing is off. Press **Ctrl + Alt + A** or click the ق icon: it turns green.

### Ctrl + Alt + A does something else in my app

Change the shortcut: right-click **ق** → **Shortcuts**. See [Shortcuts](shortcuts.md).

### Nothing types in one particular window

That window is probably running **as administrator** (for example an admin Command Prompt or some installers). Windows doesn't let normal apps type into it. Type there with Qalam off, or run that app normally.

### "Windows protected your PC" (SmartScreen)

Click **More info → Run anyway**. It appears once, for the downloaded installer. See [Install on Windows](install-windows.md).

### The setup page says "Qalam isn't answering"

The page talks to the running Qalam. If you quit Qalam or restarted the PC, open the page again from the tray: right-click **ق** → **Setup & settings…** (each start gives the page a new address).

### اللَّه shows a doubled shadda or alif

Qalam fixes this by itself in Chrome, Edge and apps built on them. If you see it in another app, [report it](https://github.com/SaShakib/qalam/issues/new?template=bug_report.yml) with the app's name.

## Both

### The harakat look wrong or collide

Try another font in that app; see [Fonts](fonts.md). Noto Naskh Arabic and Scheherazade New show harakat clearly.

### A word comes out wrong

Check the spelling in the [key chart](../LAYOUT.md), and try the other versions in the options panel (↑ ↓). If Qalam is still wrong, [report it](https://github.com/SaShakib/qalam/issues/new?template=bug_report.yml) with what you typed and what you expected. Each report becomes a test case.

### Still stuck?

[Ask on the questions board](https://sashakib.github.io/qalam/questions.html).
