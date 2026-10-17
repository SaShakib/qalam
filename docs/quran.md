# Qur'an: insert, read and search

Qalam carries the whole Qur'an (Uthmani script, full harakat, from the [Tanzil Project](https://tanzil.net)) with
three translations: **English: Saheeh International**, **Bangla: Muhiuddin Khan** and **Bangla: Zohurul Hoque**.
More can be downloaded in the app. Mac only for now.

## Insert āyāt while typing (Control-Shift-Q)

1. With the Qalam keyboard on, press **⌃⇧Q** (Control-Shift-Q) in any app. A box opens at the cursor.
2. Type what you want. The dropdown follows as you type:

   | You type | You get |
   |---|---|
   | `2:255` | Āyat al-Kursī |
   | `2:255-257` | three āyāt |
   | `baqarah 255`, `al imran 3-5`, `yaseen:1` | the sūrah by name (spelling is loose) |
   | `baq`, `cow`, `36` | the sūrah; pick it, then type the āyah number |
   | `112` + Return | the whole sūrah |
   | `rahmaan`, `qul huwa allaah` | search: Latin is read the Qalam way (`H` = ح), similar letters count too |
   | `الرحمن` or `ٱلرَّحْمَٰنِ` | search, harakat optional |

3. **↑ / ↓** choose, **Return** inserts, **Esc** closes.
4. **Tab** changes what goes in: **Arabic only**, **Arabic + translation** or **Translation only**.
   **⇧Tab** switches the translation. Qalam remembers both.

Each āyah ends with its reference, like `(2:255)`. With a translation, the Arabic line comes first and the
translation under it.

The shortcut can be changed in **Qalam → Settings → Shortcuts**. The ق menu has **Insert from the Qur'an…** too.

The box shows the Arabic in the Qur'an font; what goes into your app takes that app's font. For the
Qur'an font in Pages, Word or Notes, use **Copy** on the app's Qur'an page (it copies rich text).

## The Qur'an page in the app

**Qalam → Qur'an**:

- Every sūrah, with the translation under each āyah. Choose a range (Āyāt *from* … *to*) and **Copy** it.
- **Search the whole Qur'an** at the top: Arabic with or without harakat, or Latin letters. Type `2:255` and
  Return to jump there.
- **Translation**: pick any translation on this Mac. **More translations** downloads others from tanzil.net
  (Pickthall, Yusuf Ali, Hilali & Khan, Wahiduddin Khan, Urdu, Hindi, Indonesian, Turkish, French…).
- **⌃⇧Q inserts**: the same choice as Tab in the box.

## From the command line (for scripts and AI tools)

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
