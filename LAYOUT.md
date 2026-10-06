# Qalam: Key Layout, v2 ("spell it")

**The idea:** type a word the way you spell it in Arabic, **letter, then its haraka, letter, then its haraka**.

```
kaf-fatha  alif  ta-fatha  ba-fatha       →   k a  A  t a  b a   →   kaAtaba   →   كَاتَبَ
```

You never type sukūn or shadda; they appear by themselves. A letter with no haraka after it gets ْ, and a doubled letter gets ّ.

---

## 1. Letters (Arabic alphabet order)

| Letter | Key | Example | Result |
|---|---|---|---|
| ا | `A` (or `aa`) | `kaAtaba` / `kaataba` | كَاتَبَ |
| ب | `b` | `baAbuN` | بَابٌ |
| ت | `t` | `tamruN` | تَمْرٌ |
| ث | `th` | `thawbuN` | ثَوْبٌ |
| ج | `j` | `jabaluN` | جَبَلٌ |
| ح | `H` | `HubbuN` | حُبٌّ |
| خ | `kh` | `khubzuN` | خُبْزٌ |
| د | `d` | `daAruN` | دَارٌ |
| ذ | `dh` | `dhahabuN` | ذَهَبٌ |
| ر | `r` | `rajuluN` | رَجُلٌ |
| ز | `z` (or `zh`) | `zaytuN` | زَيْتٌ |
| س | `s` | `salaAmuN` | سَلَامٌ |
| ش | `sh` | `shamsuN` | شَمْسٌ |
| ص | `S` | `SabruN` | صَبْرٌ |
| ض | `D` | `DaAlluN` | ضَالٌّ |
| ط | `T` | `TaAlibuN` | طَالِبٌ |
| ظ | `Z` | `ZulmuN` | ظُلْمٌ |
| ع | `e` | `eilmuN` | عِلْمٌ |
| غ | `g` (or `gh`) | `gafara` | غَفَرَ |
| ف | `f` | `fajruN` | فَجْرٌ |
| ق | `q` | `qalamuN` | قَلَمٌ |
| ك | `k` | `kalbuN` | كَلْبٌ |
| ل | `l` | `layluN` | لَيْلٌ |
| م | `m` | `maAAuN` | مَاءٌ |
| ن | `n` | `nuwruN` | نُورٌ |
| ه | `h` | `huwa` | هُوَ |
| و | `w` | `waladuN` | وَلَدٌ |
| ي | `y` | `yawmuN` | يَوْمٌ |
| ء | `'`, or `A` + haraka | see §3 | |
| ة | `t'` (t with a mark) | `madrasat'uN` | مَدْرَسَةٌ |
| ى | `Y` | `ealaY` | عَلَى |
| آ | `AA` | `AAmana` | آمَنَ |

**Capitals are the heavy letters:** `H` ح · `S` ص · `D` ض · `T` ط · `Z` ظ. Lowercase gives the light one.

## 2. Harakat

| Haraka | Key | Example | Result |
|---|---|---|---|
| fatḥa ـَ | `a` | `kataba` | كَتَبَ |
| kasra ـِ | `i` | `bismi` | بِسْمِ |
| ḍamma ـُ | `u` | `kutubuN` | كُتُبٌ |
| tanwīn ـٌ ـٍ ـً | `uN` `iN` `aN` | `kitaAbuN` `kitaAbiN` `kitaAbaN` | كِتَابٌ كِتَابٍ كِتَابًا |
| shadda ـّ | type the letter twice | `rabbi` | رَبِّ |
| sukūn ـْ | nothing (automatic) | `qul` | قُلْ |

**Long vowels:** spell them with the letter, or double the vowel. Both give the same result.

| Spell it | or | Result |
|---|---|---|
| `aA` | `aa` | ـَا |
| `iy` | `ii` | ـِي |
| `uw` | `uu` | ـُو |

So with any letter: `eaA` = `eaa` (عَا), `eiy` = `eii` (عِي), `euw` = `euu` (عُو), `kaAtaba` = `kaataba`, `fiy` = `fii`. The app shows the second spelling under each word ("or …").

A voweled hamza also has two spellings: `saAala` = `sa'ala`, `qaraAa` = `qara'a`.

`ay` and `aw` are diphthongs: `bayt` → بَيْتْ, `yawm` → يَوْمْ.

---

## 3. The three tricky ones: alif, hamza, ʿayn

### The one rule for `A`

> **An alif can never carry a haraka.** So **`A` followed by a haraka is a hamza**, and **`A` followed by anything else is the alif letter.**

| You mean | Type | Result |
|---|---|---|
| kāf-fatḥa, **alif**, tāʾ-fatḥa, bāʾ-fatḥa | `kaAtaba` | كَاتَبَ |
| kāf-fatḥa, **hamza-fatḥa**, tāʾ-fatḥa, bāʾ | `kaAatab` | كَأَتَبْ |
| qāf-fatḥa, rāʾ-fatḥa, **hamza-fatḥa** | `qaraAa` (or `qara'a`) | قَرَأَ |
| jīm-fatḥa, **alif**, **hamza-fatḥa** | `jaAAa` / `jaaAa` / `jaa'a` | جَاءَ |
| **alif-madda** | `AAmana` · `qurAAnuN` | آمَنَ · قُرْآنٌ |

### `'` = hamza with no haraka (a "stopped" hamza)

`A` needs a haraka after it, so a hamza with sukūn is typed with `'`. (`'` also works for a voweled hamza if you prefer it.)

| Type | Result |
|---|---|
| `la'` | لَأْ |
| `laA'` (or `laa'`) | لَاءْ |
| `lu'luAuN` (or `lu'lu'uN`) | لُؤْلُؤٌ |
| `mu'minuN` | مُؤْمِنٌ |
| `bi'ruN` | بِئْرٌ |
| `shay'uN` | شَيْءٌ |

**You never choose the seat** (أ إ ؤ ئ ء). The engine follows the standard rule: the stronger vowel wins, **kasra > ḍamma > fatḥa > sukūn**. So `suAaAluN` → سُؤَالٌ, `raAiysuN` → رَئِيسٌ, `yasAalu` → يَسْأَلُ.

### ʿAyn = `e` (as in **E**id → عِيد)

ʿAyn is a full letter with its own key, so it never gets mixed up with hamza or alif:

| You mean | Type | Result |
|---|---|---|
| lām-fatḥa, **ʿayn** | `lae` | لَعْ |
| lām-fatḥa, **hamza** | `la'` | لَأْ |
| lām-fatḥa, alif, **ʿayn** | `laAe` | لَاعْ |
| lām-fatḥa, alif, **hamza** | `laA'` | لَاءْ |
| **asked** (hamza) | `saAala` | سَأَلَ |
| **coughed** (ʿayn) | `saeala` | سَعَلَ |
| ʿīd | `eiyduN` | عِيدٌ |
| maʿa | `maea` | مَعَ |
| naʿam | `naeam` | نَعَمْ |
| taʿallama | `taeallama` | تَعَلَّمَ |

The last pairs show why `a'` couldn't work for ʿayn: `sa'ala` would have to be both سَأَلَ (asked) and سَعَلَ (coughed).

---

## 4. Start of a word: waṣl or qaṭʿ

| Starts with | Means | Type | Result |
|---|---|---|---|
| a bare vowel | hamzat **al-waṣl** | `ismu` · `uktub` | اِسْمُ · اُكْتُبْ |
| `A` + haraka | hamzat **al-qaṭʿ** | `AaHmadu` · `Aislaamu` · `Aummu` | أَحْمَدُ · إِسْلَامُ · أُمُّ |
| `al` | the article | `alkitaAbu` · `alshamsu` · `alladhiy` | الْكِتَابُ · الشَّمْسُ · الَّذِي |

Sun letters get shadda and leave the lām bare. Moon letters put sukūn on the lām. `wal-`, `fal-`, `bil-`, `kal-`, `lil-` work the same way: `wal-qamari` → وَالْقَمَرِ, `lil-naAsi` → لِلنَّاسِ. In **Qur'an style**, waṣl becomes ٱ: ٱسْمُ، ٱلْكِتَابُ، وَٱلْقَمَرِ.

## 5. Tā' marbūṭa ة = `t'`

Think of it as "t with a mark". It only ever comes at the end of a word. A hamza after ت is still easy to type: `tAa`, `tAu`… (the `A` rule), so `t'` never clashes.

| Type | Result |
|---|---|
| `madrasat'` | مَدْرَسَة |
| `madrasat'uN` | مَدْرَسَةٌ |
| `madrasat'aN` | مَدْرَسَةً (no extra alif after ة) |
| `SalaAt'` | صَلَاة |

## 6. Spelling words (typed as heard, spelled as written)

| Type | Everyday | Qur'an style |
|---|---|---|
| `allaAhu` | اللَّهُ | ٱللَّهُ |
| `lillaAhi` | لِلَّهِ | لِلَّهِ |
| `haAdhaA` / `haAdhihi` | هَٰذَا / هَٰذِهِ | هَٰذَا / هَٰذِهِۦ |
| `dhaAlika` | ذَٰلِكَ | ذَٰلِكَ |
| `laAkin` / `laAkinna` | لَٰكِنْ / لَٰكِنَّ | لَٰكِنْ / لَٰكِنَّ |
| `AilaAhuN` | إِلَٰهٌ | إِلَٰهٌ |
| `AulaA'ika` | أُولَٰئِكَ | أُو۟لَٰٓئِكَ |
| `alraHmaAnu` | الرَّحْمَٰنُ | ٱلرَّحْمَٰنُ |

## 7. Symbol keys

| Key | Does |
|---|---|
| `` ` `` | separator: `s`h` = سه, not ش |
| `-` | joiner, prints nothing: `wa-qaAla` → وَقَالَ |
| `^` | dagger alif ـٰ: `alea^lamiyna` → الْعَٰلَمِينَ (Qur'an: ٱلْعَٰلَمِينَ) |
| `=` | maddah ـٓ (Qur'an): `jaA='a` → جَآءَ |
| `~` | shadda by hand (rarely needed) |
| `o` | sukūn by hand (only in *As typed* mode) |
| `_` | taṭwīl ـ |
| `\` | Qur'anic marks, waqf signs, āyah numbers, phrases (unchanged from v1: `\j` ۚ, `\7` ﴿٧﴾, `\bism`, `\saw` ﷺ …) |
| `,` `;` `?` | ، ؛ ؟ |

## 8. Try-it sentences

| Type | Result |
|---|---|
| `dhahaba alwaladu AilaY almadrasat'i.` | ذَهَبَ الْوَلَدُ إِلَى الْمَدْرَسَةِ. |
| `AanaA AuHibbu allugat'a alearabiyyat'a` | أَنَا أُحِبُّ اللُّغَةَ الْعَرَبِيَّةَ |
| `hal fahimta?` | هَلْ فَهِمْتَ؟ |
| *(Qur'an)* `qul huwa allaAhu AaHaduN` | قُلْ هُوَ ٱللَّهُ أَحَدٌ |
| *(Qur'an)* `alHamdu lillaAhi rabbi alea^lamiyna` | ٱلْحَمْدُ لِلَّهِ رَبِّ ٱلْعَٰلَمِينَ |
