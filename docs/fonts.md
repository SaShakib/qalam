# Fonts

Qalam types standard Unicode Arabic, so **the app you type in decides the font**. Any Arabic font shows the text; some show harakat more clearly than others.

## The default fonts

| Text | Font | Notes |
|---|---|---|
| Everyday Arabic | **Scheherazade New** (SIL) | large, very readable harakat; included with Qalam |
| Qur'an text | **KFGQPC HAFS Uthmanic Script** (King Fahd Complex) | the standard Madinah mushaf font; included with Qalam |

On the Mac you can pick other fonts in **Qalam → Settings → Fonts in this app** (also Noto Naskh Arabic, Amiri Quran and Noto Sans Arabic, which are included), and change the Arabic text size. On Windows, Qalam's typing box uses Scheherazade New.

## Fonts in all your apps

Qalam types standard Unicode Arabic, so **the app you type in decides the font**. Qalam installs Scheherazade New and KFGQPC Hafs for every app (Word, Pages, Notes, your browser…), just for your user, with no password:

| Installed with | The fonts |
|---|---|
| Mac installer package (`.pkg`) or the Terminal command | installed automatically |
| Homebrew | click **Install fonts for all apps** in the setup's Fonts step, or in Settings → Fonts |
| Windows installer (`Qalam-Setup`) | installed automatically |
| Windows single `.exe` | click **Install fonts for all apps** in the setup page's Fonts step (right-click ق → Setup & settings…) |

Then choose the font in your app.

> **In Microsoft Word,** Arabic uses the **Complex scripts** font: Home → Font dialog (Ctrl + D) → **Complex scripts**.

If you'd rather install fonts yourself, they're in the [`fonts/` folder](https://github.com/SaShakib/qalam/tree/main/fonts) of the repository with their licences. KFGQPC Hafs comes from the [King Fahd Complex](https://fonts.qurancomplex.gov.sa/); its licence lets Qalam include it unmodified and free of charge. On a Mac, double-click a `.ttf` file → **Install**; on Windows, right-click → **Install**.

## Harakat look wrong?

- **Doubled shadda or alif on اللَّه in Chrome-based apps** (Chrome, Edge, VS Code, Slack, WhatsApp…): Qalam fixes this by itself in those apps, on Mac and Windows. Text you copy from the Mac app's **Write** page has a **Copy for browsers & chat apps** option for the same fix.
- **Madinah-mushaf sukūn (ۡ) shows as a box:** the font doesn't have it. Use KFGQPC Hafs, or switch the setting off.
