# Qur'an: insert, read and search

Qalam carries the whole Qur'an (Uthmani script, full harakat, from the [Tanzil Project](https://tanzil.net)) with
translations: **English**: Saheeh International, Mufti Taqi Usmani; **Bangla**: Muhiuddin Khan, Bayaan Foundation,
Sheikh Mujibur Rahman, Taisirul Quran (Tawheed Publication), Dr. Abu Bakr Zakaria, Zohurul Hoque; **Urdu**: Mufti Taqi Usmani.
On the Mac more can be downloaded in the app. Windows has the insert box and the Qur'an page with all nine bundled translations.

## Insert āyāt while typing (Mac: Control-Shift-Q · Windows: Ctrl+Alt+Q)

**Mac.** Press **⌃⇧Q** in any app, with any keyboard on (Qalam stays in the menu bar as ق and starts at
login for this; switch it off in **Settings → Qur'an box**). A box opens:

1. **Search** at the top. It searches the **Arabic** by default: type part of an āyah the Qalam way
   (`qul huwa allaahu`, `inna allaaha maea alSSaAbiriyna`) and the field shows it in Arabic as you type
   (Backspace removes the last key). Harakat are optional, words the mushaf joins (يَٰٓأَيُّهَا) match when
   typed apart, and pasted Arabic works too. `2:255`, `baqarah 255-257` and sūrah names work as well.
   Click **Translation** next to the field to search the translation's words instead (English, or Bangla
   letters for the Bangla one). The matched words are highlighted in each result.
2. Choose a result with **↑ / ↓** or a **click** (double-click inserts). It fills **Sūrah**, **From āyah** and
   **To āyah**.
3. Set the range with the **From āyah** / **To āyah** dropdowns. The preview shows exactly what goes in.
4. Choose **Arabic only**, **Arabic + translation** or **Translation only**, and the translation. Pick the
   Qur'an font and, if you are pasting into a rich editor, turn on **Styled paste** or **Paste as artwork**.
5. **Insert** (Return). **Esc** clears the search, then closes the box.

The text goes back into the app you were in. With the Qalam keyboard on there, it is typed in directly.
With another keyboard, Qalam pastes it (⌘V) once you allow it in **System Settings → Privacy & Security →
Accessibility** (the box has an **Allow…** button); until then it is copied and you press ⌘V.
The setup's **Qur'an box** step (also in **Settings → Qur'an box**) shows both permissions with a ✓ when
they're on. After a Qalam update macOS forgets the Accessibility permission (it belongs to the exact build);
**Allow…** clears the old entry so you can switch it on again.

**Windows.** Press **Ctrl+Alt+Q** (Qalam running in the tray, Arabic on or off). The same box as on the Mac
opens in the middle of the screen, on the last āyāt you inserted:

1. **Search** at the top, **Arabic** by default: type part of an āyah the Qalam way (`qul huwa allaahu`) and
   the field shows it in Arabic (Backspace removes the last key, Ctrl+Backspace empties the field). Harakat
   are optional, and words the mushaf joins (يَٰٓأَيُّهَا) match when typed apart. `2:255`,
   `baqarah 255-257` and sūrah names work too. Click **Translation** next to the field to search the
   translation's words instead (the chosen one if it's English, otherwise Saheeh International).
2. The results show under the field (five rows, then the list scrolls with ↑ / ↓ or the mouse wheel). The
   chosen one fills **Sūrah**, **From āyah** and **To āyah**; a double-click inserts it.
3. Set the range with the dropdowns (a click opens a menu; or **Tab** to them, see below). The preview
   shows the āyāt in the Qur'an font, with the translation under each when it goes in too.
4. At the bottom choose **Arabic only**, **Arabic + translation** or **Translation only**, and the translation.
   Choose a **Qur'an font** and optionally set **Styled paste: On** or **Paste as artwork: On**.
5. **Insert** (Enter). **Esc** clears the search, then closes the box.

   | Key | Does |
   |---|---|
   | letters, Space | search (Ctrl+V pastes Arabic or Bangla letters, which search the Arabic / a Bangla translation) |
   | ↑ / ↓ | choose a result |
   | Tab / Shift+Tab | go to **Sūrah**, **From āyah**, **To āyah** and back to the search |
   | on a dropdown: ↑ / ↓ | change it; type `36` or `yasin`, or `2:255-257` to fill all three; `/` back to the search |
   | Enter | insert |
   | Esc | clear the search; again: close |

The box is short at first and grows downward only when results or more āyāt need room. A long passage shows
its first āyāt and "… and N more" (all of them are inserted).

Each Arabic āyah ends with an Arabic-Indic number inside Qur'anic ornamental brackets, such as `﴿٢٥٥﴾`;
Qalam never appends a `(sūrah:āyah)` reference to the inserted Arabic. With a translation, the Arabic line
comes first and the translation under it.

The shortcut can be changed in **Qalam → Settings → Shortcuts** on the Mac (the ق menu has
**Insert from the Qur'an…** too), and on Windows in the tray menu → **Shortcuts** → **Insert from the Qur'an**
(or **Setup & settings… → Shortcuts**).

**Windows notes**

- Keys are read on the US layout, like Qalam's typing, and the box never takes the focus from your app.
  To search in Arabic or Bangla letters, copy them and press **Ctrl+V** in the box.
- A click on a dropdown opens a Windows menu (long lists in columns). The menu has the focus while it is
  open; your app gets it back as soon as you pick.
- Lines are separated with **Shift+Enter**, so in chat apps (WhatsApp, Teams, Slack) the āyah and its
  translation stay in one message instead of sending it.
- Clicking another window or pressing another Ctrl/Alt shortcut closes the box.
- In Chrome-based apps the same invisible fix as for typing keeps اللَّه from turning into the font's ligature.
- The box is sized for the screen it opens on (its DPI), and Bangla translations show in Nirmala UI.

The box defaults to **KFGQPC HAFS Uthmanic Script** and also offers Amiri Quran and Noto Naskh Arabic.
Normal insertion is plain editable Unicode, so its final appearance follows the receiving app's font. With
**Styled paste** on, Qalam also places an RTF version using the chosen font on the clipboard; Illustrator,
Word, OneNote, Notes, and rich browser editors can preserve it while leaving the text editable. A plain
text field (for example Notepad, the Chrome address bar, or a normal web text box) cannot retain any font
by design and receives the same editable Unicode. Install the chosen font from **Setup & settings… → Fonts**
on Windows before using it independently in Illustrator.

**Paste as artwork** is a separate saved option. It places a rendered image and the Unicode text on the
clipboard. Illustrator and rich Chrome editors can paste the image, making its appearance independent of
the receiving app's fonts; it is not editable as text. If a target cannot paste images, it can use the
Unicode fallback instead. Plain-text fields cannot paste graphics at all.

## The Qur'an page in the app

**Windows:** open the Qalam window → **Qur'an**: every sūrah, a translation under each āyah, search, and **Copy** for any range. **Mac:**

**Qalam → Qur'an**:

- Every sūrah, with the translation under each āyah. Choose a range (Āyāt *from* … *to*) and **Copy** it.
- **Search the whole Qur'an** at the top: Arabic with or without harakat, or Latin letters. Type `2:255` and
  Return to jump there.
- **Translation**: pick any translation on this Mac. **More translations** downloads others from tanzil.net
  (Pickthall, Yusuf Ali, Hilali & Khan, Wahiduddin Khan, Urdu, Hindi, Indonesian, Turkish, French…).
- **⌃⇧Q inserts**: the same choice as Tab in the box.

## From the command line (Mac; for scripts and AI tools)

```
qalam ayah 2:255 "baqarah 1-5" --trans en.sahih,bn.bengali
qalam ayah 2:255-257 --trans bn.hoque --insert     # exactly what ⌃⇧Q would insert
qalam search "الرحمن" --max 20 --trans none
qalam surahs
qalam translations
```

`qalam` is built by `swift build -c release --product qalam`. It finds the data in `quran/` next to the
repo, in an installed Qalam.app, or wherever `QALAM_QURAN` points.

## Sources

The Qur'an text and translations are verbatim from tanzil.net; see [quran/README.md](https://github.com/SaShakib/qalam/blob/main/quran/README.md).
