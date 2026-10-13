# FAQ

## Is Qalam free? Is it safe?

Yes, it's free and open source: all the code is [on GitHub](https://github.com/SaShakib/qalam). Qalam works offline. It never sends what you type anywhere; the only thing it does online is ask GitHub whether there is a new version.

## Why does my Mac or PC warn me when I install it?

The downloads aren't signed with a paid Apple or Microsoft certificate yet, so the first install shows one warning. [Install on Mac](install-mac.md) and [Install on Windows](install-windows.md) show what to click. Updates after that don't warn again. Homebrew and the Terminal command never show the warning on a Mac.

## Do I need to know the Arabic keyboard layout?

No. If you know how a word is spelled, letter by letter with its harakat, you can type it. See [Typing basics](typing.md).

## How do I type a hamza?

Type `A` followed by a haraka (`saAala` → سَأَلَ), or `'` for a hamza with no haraka. Qalam picks the seat (أ إ ؤ ئ ء) from the vowels around it. If you want another seat, pick it in the options panel with ↑ ↓.

## Why didn't my word get a sukūn?

With **Smart** sukūn (the default), sukūn only goes on a stop *inside* a word, never on the last letter (قُل) or on ي/و after fatḥa (بَيت). Type `o` to add one by hand, or choose **Full** sukūn in [Settings](settings.md).

## Why did I get letters with no harakat at all?

You typed no vowels (`ktb` → كتب), which Qalam reads as "bare letters, please". Or Harakat is set to **None**: switch it back in [Settings](settings.md) or with the Harakat button in the options panel.

## How do I type an English word without switching keyboards?

Type it and press **Esc** before Space: Qalam gives back the English letters.

## How do I type ة, ى, ع?

`t'` → ة, `Y` → ى, `e` → ع. Everything is in the [key chart](../LAYOUT.md).

## Why is ص coming out instead of س?

A capital `S` is ص (the heavy letter). Only **Shift** makes a capital in Qalam; Caps Lock is ignored.

## Can I use it on my phone or on Linux?

The keyboard is for Mac and Windows. On other devices, the [typing pad on the website](https://sashakib.github.io/qalam/#demo) works in the browser, and you can copy the Arabic.

## Does it work in Word, Chrome, WhatsApp…?

Yes, in every app. In Chrome-based apps Qalam adds an invisible joiner inside اللَّه so its harakat draw correctly.

## How do I change a shortcut?

See [Shortcuts](shortcuts.md).

## Something else?

Look at [Troubleshooting](troubleshooting.md) and the [questions board](https://sashakib.github.io/qalam/questions.html), or [ask a question](https://github.com/SaShakib/qalam/issues/new?template=question.yml).
