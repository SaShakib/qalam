# Demo video

A 30-second video (1920×1080, 30 fps) made from an HTML animation: `qalam-demo.html`. Every Arabic word in it comes from the real Qalam engine, so it always matches what the app types.

| Time | Scene |
|---|---|
| 0–3 s | Title: قَلَم, "Type Arabic the way you spell it" |
| 3–7.5 s | The setup steps: keyboard on, style, harakat & sukūn |
| 7.5–13.5 s | Typing words letter by letter: `kataba`, `kitaAbuN`, `madrasat'uN` |
| 13.5–18.5 s | A long Qur'an-style phrase: `alHamdu lillaAhi rabbi alea^lamiyna` |
| 18.5–23 s | Suggestions: `qaraA`, ↓ ↓, Space |
| 23–27 s | Sukūn off (Ctrl+Alt+O), then harakat off |
| 27–30 s | End card: free for Mac & Windows, website address |

## Make it

```bash
make wasm                     # once: the animation uses the real engine
npm i -D playwright           # or a global Playwright
node video/render.js          # → video/out/
node video/render.js --site   # … and copy it to website/media/ for the website
```

`video/out/` then has:

| File | What |
|---|---|
| `qalam-demo.mp4` | with subtitles drawn in |
| `qalam-demo-clean.mp4` | without subtitles (for YouTube with the `.srt`, or for adding your own) |
| `qalam-demo.srt`, `qalam-demo.vtt` | the subtitles as files |
| `qalam-demo-voice.mp4` | with your voice-over, if `video/voice.mp3` exists |
| `poster.png` | a still for thumbnails |

To watch the animation live, serve the repository (`python3 -m http.server` in the repository folder) and open http://localhost:8000/video/qalam-demo.html.

To change words, timing or subtitles, edit `qalam-demo.html` (`SUBTITLES` at the top, and the scenes in `app()`) and render again.

## Voice-over

### The script

Read one line per scene. Each line fits its time slot at a calm pace (about 2.5 words a second):

| Start | Say |
|---|---|
| 0.3 s | Meet Qalam. Type Arabic the way you spell it. |
| 3.1 s | Install it, and a short setup turns on the keyboard and picks your style. |
| 7.5 s | Type each letter, then its haraka. Sukūn and shadda appear by themselves. |
| 13.5 s | Long phrases too: hamza, the article and Qur'anic marks are done for you. |
| 18.5 s | Suggestions appear as you type. Arrows choose, Space puts it in. |
| 23.0 s | Turn sukūn or harakat off with one shortcut. |
| 27.0 s | Free for Mac and Windows. Get it at sashakib dot github dot io slash qalam. |

Pronunciation: Qalam = *KA-lam*, sukūn = *su-KOON*, haraka = *HA-ra-ka*, harakat = *ha-ra-KAAT*, shadda = *SHAD-da*.

### Add it

1. Record or generate the voice (see below) as **one file, 30 seconds or a little less**, starting at 0.
2. Save it as `video/voice.mp3` (or `.wav` / `.m4a`).
3. Run `node video/render.js --site`. You get `video/out/qalam-demo-voice.mp4`, and the website uses it.

Or with ffmpeg only, on the finished video:

```bash
ffmpeg -i video/out/qalam-demo.mp4 -i voice.mp3 -map 0:v -map 1:a -c:v copy -c:a aac -af apad -t 30 qalam-demo-voice.mp4
```

### Free ways to get a voice

| Option | Cost | Good for |
|---|---|---|
| **Record yourself** with your phone, or [Audacity](https://www.audacityteam.org/) on the computer | free | the most natural, and you can say the Arabic words properly. Record each line, then line them up to the times above. |
| **[Piper](https://github.com/rhasspy/piper)** | free, open source, runs offline | good neural voices with no account and no limits. `pip install piper-tts`, then `echo "Meet Qalam…" \| piper -m en_US-lessac-medium --output_file line1.wav` |
| **[CapCut](https://www.capcut.com/)** (desktop or web) | free | import the video, use **Text to speech** on each subtitle, and it can also make auto captions. Easiest all-in-one editor. |
| **[Clipchamp](https://clipchamp.com/)** (built into Windows 11) | free tier | **Record & create → Text to speech**, many voices and languages, export 1080p |
| **[edge-tts](https://github.com/rany2/edge-tts)** | free | Microsoft's neural voices from the command line: `edge-tts --voice en-US-AriaNeural --text "…" --write-media line1.mp3`. Check Microsoft's terms for your use. |
| **[ElevenLabs](https://elevenlabs.io/)** | free tier (about 10,000 characters a month) | very natural voices. The free tier needs attribution and is for non-commercial use. |
| **[TTSMaker](https://ttsmaker.com/)** | free | quick web text-to-speech with MP3 download |

Tip for generated voices: make one audio file per line, then put them together at the start times above, in CapCut, Clipchamp or Audacity. That keeps the voice in step with the subtitles.
