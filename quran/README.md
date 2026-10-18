# Qur'an data

Used by the keyboard's **Insert from the Qur'an** box (Mac ⌃⇧Q, Windows Ctrl+Alt+Q; Windows embeds a gzipped copy, `windows/assets/quran`, made by `make guide`), the app's **Qur'an** page and
`qalam ayah 2:255` / `qalam search …` on the command line.

| File | What | Source |
|---|---|---|
| `quran-uthmani.txt` | The Qur'an, Uthmani script with full harakat (Hafs), one āyah per line `sura|aya|text` | [Tanzil Project](https://tanzil.net) |
| `en.sahih.txt` | English: Saheeh International | Tanzil translations |
| `bn.bengali.txt` | Bangla: Muhiuddin Khan | Tanzil translations |
| `bn.hoque.txt` | Bangla: Zohurul Hoque | Tanzil translations |
| `bn.bayaan.txt` | Bangla: Bayaan Foundation (Rawai Al-bayan) | quran.com API (translation 162), footnotes left out |
| `bn.mujibur.txt` | Bangla: Sheikh Mujibur Rahman (Darussalaam) | quran.com API (163) |
| `bn.taisirul.txt` | Bangla: Taisirul Quran (Tawheed Publication) | quran.com API (161) |
| `bn.zakaria.txt` | Bangla: Dr. Abu Bakr Muhammad Zakaria | quran.com API (213), footnote numbers left out |
| `en.usmani.txt` | English: Mufti Taqi Usmani | quran.com API (84) |
| `ur.usmani.txt` | Urdu: Mufti Muhammad Taqi Usmani | github.com/fawazahmed0/quran-api |
| `surahs.tsv` | Sūrah names and āyah counts | Tanzil metadata (CC BY 3.0) |
| `translations.tsv` | Translations the app can show; `no` = downloaded from tanzil.net on request | |

The text files are verbatim copies from tanzil.net, with their copyright notices at the end
(Tanzil's terms: the Qur'an text may be used in any application if the source is named and linked,
and it must not be changed). Tanzil puts the basmala before āyah 1 of each sūrah; Qalam leaves it out
when it shows or inserts āyah 1 (except Al-Fātiḥa, where it is āyah 1).
The translations are for non-commercial use, under their translators' terms as given on tanzil.net.

Translations the app downloads are kept in `~/Library/Application Support/Qalam/Quran/`.
