# Demo video

A narrated walkthrough of Qalam (about 2¼ minutes, 1920×1080, 30 fps) made from an HTML animation, `qalam-demo.html`. Every Arabic word in it comes from the real Qalam engine.

The voice drives the timing: **each scene lasts as long as its narration**, so picture, voice and subtitles always match, whatever the voice's pace.

## Scenes

| # | Scene | Narration (see `script.json`) |
|---|---|---|
| 1 | Title | What Qalam is |
| 2 | Setup steps | The guided setup after installing |
| 3 | Switch keyboards | Control-Space on Mac, Ctrl+Alt+A on Windows |
| 4 | Typing | Letter then vowel; sukūn and shadda by themselves |
| 5 | Unusual letters | e = ع, H S D T Z, th dh kh gh, t' = ة, Y = ى |
| 6 | Hamza | Capital A + vowel; the seat is chosen |
| 7 | Phrases | The article, sun letters, Qur'an style |
| 8 | Suggestions | The options panel, ↑ ↓ and Space |
| 9 | Sukūn and harakat | Smart sukūn, the shortcuts, the Harakat button |
| 10 | Shortcuts | Changing a shortcut in Settings |
| 11 | Where the settings are | The Mac ق menu, the Windows tray menu |
| 12 | Fonts | Font and size in Settings |
| 13 | Updates | Settings → Updates; Windows tray menu |
| 14 | End card | Free, offline, private; the website |

## Make it

```bash
make wasm                                   # once: the animation uses the real engine
ELEVENLABS_API_KEY=… node video/elevenlabs.js   # the voice: video/voice/01.mp3 … 14.mp3
node video/render.js --site                 # the video, and a copy on the website
```

`elevenlabs.js` picks an American male voice from your ElevenLabs account (it chose **Brian**) and reads slightly slower than normal (`speed` in `script.json`). Pass a voice id to use another voice. `--only 5,9` remakes just those lines.

`video/out/` then has:

| File | What |
|---|---|
| `qalam-demo.mp4` | with voice and subtitles: **post this one** |
| `qalam-demo-clean.mp4` | with voice, no subtitles (for YouTube, with the `.srt`) |
| `qalam-demo.srt`, `qalam-demo.vtt` | the subtitles |
| `poster.png` | a thumbnail |

To watch the animation live (with estimated timing and no voice): run `python3 -m http.server` in the repository folder and open http://localhost:8000/video/qalam-demo.html.

## Change it

- **Words:** edit `script.json`. `say` is what the voice reads (spelled for good pronunciation, such as "sukoon"); `show` is the subtitle. Then remake that line's voice (`--only N`) and render again.
- **Pictures:** each scene is an entry in `SCENES` in `qalam-demo.html`, with a `build` and a `draw(el, u, len)`; timings inside a scene are fractions of its length, or tied to its subtitle sentences.

## Posting

On ElevenLabs' free plan the voice is for non-commercial use, and you must credit ElevenLabs. Put something like this in the description:

> Qalam: type Arabic the way you spell it. Free for Mac & Windows.
> Download: https://sashakib.github.io/qalam/
> Voice: ElevenLabs
