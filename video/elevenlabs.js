// Makes the voice-over with ElevenLabs: one clip per subtitle line, into video/voice/01.mp3 … 07.mp3.
// Then `node video/render.js --site` lines them up with the subtitles and puts the video on the website.
//
//   ELEVENLABS_API_KEY=… node video/elevenlabs.js [voice-id]
//
// The key is in ElevenLabs → your profile → API Keys. Without a voice id it uses "Rachel" (a calm
// narrator voice in the default library). Free plan: attribute ElevenLabs where you post the video,
// and non-commercial use only.
const fs = require("fs");
const path = require("path");

const KEY = process.env.ELEVENLABS_API_KEY;
const VOICE = process.argv[2] || process.env.ELEVENLABS_VOICE_ID || "21m00Tcm4TlvDq8ikWAM";
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
