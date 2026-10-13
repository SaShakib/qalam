// Makes the voice-over with ElevenLabs: one clip per subtitle line, into video/voice/01.mp3 … 07.mp3.
// Then `node video/render.js --site` lines them up with the subtitles and puts the video on the website.
//
//   ELEVENLABS_API_KEY=… node video/elevenlabs.js [voice-id] [--only 7]   (--only: just those lines, e.g. 3,7)
//
// The key is in ElevenLabs → your profile → API Keys. Without a voice id it uses "Alice" (a clear
// educator voice, available on the free plan). Free plan: attribute ElevenLabs where you post the video,
// and non-commercial use only.
const fs = require("fs");
const path = require("path");

const KEY = process.env.ELEVENLABS_API_KEY;
const args = process.argv.slice(2);
const onlyAt = args.indexOf("--only");
const ONLY = onlyAt >= 0 ? args.splice(onlyAt, 2)[1].split(",").map(Number) : null;
const VOICE = args[0] || process.env.ELEVENLABS_VOICE_ID || "Xb7hH8MSUJpSbSDYk0k2";
if (!KEY) { console.error("Set ELEVENLABS_API_KEY first."); process.exit(1); }

// The subtitles are the script; a few words are spelled out the way they should be said.
const html = fs.readFileSync(path.join(__dirname, "qalam-demo.html"), "utf8");
const lines = [...html.matchAll(/^\s*\[[\d.]+, [\d.]+, "(.*)"\],\s*$/gm)].map((m) => m[1]);
const SAY = [
  [/sashakib\.github\.io\/qalam/g, "sashakib dot github dot io, slash qalam"],
  [/Sukūn/g, "Sukoon"], [/sukūn/g, "sukoon"],
];

(async () => {
  const out = path.join(__dirname, "voice");
  fs.mkdirSync(out, { recursive: true });
  for (const [i, line] of lines.entries()) {
    if (ONLY && !ONLY.includes(i + 1)) continue;
    const text = SAY.reduce((t, [re, s]) => t.replace(re, s), line);
    const r = await fetch(`https://api.elevenlabs.io/v1/text-to-speech/${VOICE}?output_format=mp3_44100_128`, {
      method: "POST",
      headers: { "xi-api-key": KEY, "Content-Type": "application/json", Accept: "audio/mpeg" },
      body: JSON.stringify({ text, model_id: "eleven_multilingual_v2", voice_settings: { stability: 0.5, similarity_boost: 0.75, style: 0.2 } }),
    });
    if (!r.ok) throw new Error(`line ${i + 1}: ElevenLabs answered ${r.status} ${await r.text()}`);
    const file = path.join(out, `${String(i + 1).padStart(2, "0")}.mp3`);
    fs.writeFileSync(file, Buffer.from(await r.arrayBuffer()));
    console.log(file, "←", text);
  }
  console.log("Now run: node video/render.js --site");
})().catch((e) => { console.error(e.message); process.exit(1); });
