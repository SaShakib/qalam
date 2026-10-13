# Settings

## Where to find them

| | Mac | Windows |
|---|---|---|
| All settings | **Qalam app → Settings** | Right-click **ق** in the tray → **Setup & settings…** → page with every setting |
| Quick switches | Click **ق** in the menu bar (the input menu) | Right-click **ق** in the tray |
| While typing | **Sukūn** and **Harakat** buttons in the options panel, and their [shortcuts](shortcuts.md) | the same |

Changes apply straight away, in every app.

## Typing settings (both)

| Setting | Choices | What it does |
|---|---|---|
| **Writing style** | Everyday · Qur'an (Uthmani) | Qur'an style adds ٱ, dagger alif, maddah, silent-letter marks, waqf signs and āyah numbers. |
| **Harakat** | Full · Only what I type · None | **Full** writes every haraka and adds sukūn. **Only what I type** writes only the marks you typed. **None** writes plain letters (good for chat and dictionary search). |
| **Sukūn** | Smart · Full · Off | **Smart** (default): only at a stop inside a word (مَكْتَب، قُل، بَيت); `o` adds one by hand. **Full**: every stop, like a mushaf (Qur'an style always uses this unless Off). **Off**: none at all, not even from `o`. |
| **Show options while typing** | on · off | The panel with up to 4 versions of the word. See [Typing basics](typing.md#options-while-typing). |
| **Smart spelling** | on · off | Words written differently from how they sound: اللَّه، إِلَٰه، هَٰذَا، ذَٰلِكَ، لَٰكِنْ، الرَّحْمَٰن… |
| **Arabic digits** | on · off | ١٢٣ instead of 123. |
| **Madinah-mushaf sukūn** | on · off | In Qur'an style, قۡ instead of قْ. Needs a mushaf font such as KFGQPC Hafs. |
| **Forget what I picked** (Mac) | button | Clears the versions Qalam remembered from the options panel. |

## Mac only

| Setting | Where | What it does |
|---|---|---|
| **Shortcuts** | Settings → Shortcuts | See [Shortcuts](shortcuts.md). |
| **Fonts** | Settings → Fonts in this app | The Arabic and Qur'an fonts the app uses, and the text size. See [Fonts](fonts.md). |
| **Updates** | Settings → Updates | Check automatically, install automatically, check now. See [Updates](updates.md). |
| **Keyboard status** | Settings → Keyboard status | Shows whether the keyboard is installed and in the input menu. **Add Qalam now**, **Open Keyboard Settings…**, **Run setup again**, and the 20-word practice again. |

## Windows only

| Setting | Where | What it does |
|---|---|---|
| **Arabic typing** | tray menu, or **Ctrl + Alt + A** | On or off. The tray icon is green when on. |
| **Shortcuts** | tray menu → Shortcuts, or the settings page | See [Shortcuts](shortcuts.md). |
| **Start with Windows** | tray menu, or the settings page | Starts Qalam when you sign in. |
| **Install updates automatically** | tray menu, or the settings page | See [Updates](updates.md). |
| **Check for updates…** | tray menu | Checks now and shows the version you have. |
| **Letters, words & practice…** | tray menu | The key chart and the 20-word practice, in your browser. |
| **Uninstall Qalam…** | tray menu | See [Uninstall](uninstall.md). |

## Where settings are stored

- **Mac:** `~/Library/Preferences/com.qalam.shared.plist` (keyboard) and `com.qalam.app.plist` (app).
- **Windows:** `%APPDATA%\Qalam\settings.json`.

Nothing is sent anywhere. The versions you picked in the options panel are stored in the same place.
