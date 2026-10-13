// Makes the voice-over with ElevenLabs: one clip per scene in video/script.json, into video/voice/01.mp3, 02.mp3 …
// Then `node video/render.js --site` builds the video around them: each scene lasts as long as its narration.
//
//   ELEVENLABS_API_KEY=… node video/elevenlabs.js [voice-id] [--only 3,7]
//
// Without a voice id it picks an American male voice from your account (free-plan voices work), and
// says which. The key is in ElevenLabs → your profile → API Keys. Free plan: credit ElevenLabs where
// you post the video, and non-commercial use only.
const fs = require("fs");
const path = require("path");

const KEY = process.env.ELEVENLABS_API_KEY;
if (!KEY) { console.error("Set ELEVENLABS_API_KEY first."); process.exit(1); }
const args = process.argv.slice(2);
const onlyAt = args.indexOf("--only");
const ONLY = onlyAt >= 0 ? args.splice(onlyAt, 2)[1].split(",").map(Number) : null;
const script = JSON.parse(fs.readFileSync(path.join(__dirname, "script.json"), "utf8"));
const OUT = path.join(__dirname, "voice");
const api = (p, init = {}) => fetch("https://api.elevenlabs.io" + p, { ...init, headers: { "xi-api-key": KEY, ...(init.headers || {}) } });

// Narrators we prefer, best first: clear American male voices from ElevenLabs' default set.
const PREFERRED = ["Brian", "Eric", "Bill", "Chris", "Will", "Roger", "Liam"];

async function pickVoice() {
  const saved = path.join(OUT, "voice.json");
  if (args[0]) return { voice_id: args[0], name: args[0] };
  if (ONLY && fs.existsSync(saved)) return JSON.parse(fs.readFileSync(saved, "utf8"));
  const r = await api("/v1/voices");
  if (!r.ok) throw new Error(`listing voices: ${r.status} ${await r.text()}`);
  const voices = (await r.json()).voices || [];
  const label = (v, k) => String((v.labels || {})[k] || "").toLowerCase();
  const male = voices.filter((v) => label(v, "gender") === "male");
  const american = male.filter((v) => label(v, "accent").includes("americ"));
  const pool = american.length ? american : male;
  const pick = PREFERRED.map((n) => pool.find((v) => v.name.split(/[\s-]/)[0] === n)).find(Boolean) || pool[0];
  if (!pick) throw new Error("No male voice on this account. Pass a voice id: node video/elevenlabs.js <voice-id>");
  console.log("voices on the account:", voices.map((v) => `${v.name} (${label(v, "gender")}, ${label(v, "accent")})`).join("; "));
  return { voice_id: pick.voice_id, name: pick.name };
}

(async () => {
  fs.mkdirSync(OUT, { recursive: true });
  const voice = await pickVoice();
  console.log(`voice: ${voice.name} (${voice.voice_id})`);
  if (!ONLY) for (const f of fs.readdirSync(OUT)) if (/^\d+\.(mp3|wav|m4a)$/.test(f)) fs.rmSync(path.join(OUT, f));
  fs.writeFileSync(path.join(OUT, "voice.json"), JSON.stringify(voice, null, 2) + "\n");
  const { speed, stability, similarity_boost, style } = script.voice;
  for (const [i, scene] of script.scenes.entries()) {
    if (ONLY && !ONLY.includes(i + 1)) continue;
    const r = await api(`/v1/text-to-speech/${voice.voice_id}?output_format=mp3_44100_128`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "audio/mpeg" },
      body: JSON.stringify({
        text: scene.say,
        model_id: "eleven_multilingual_v2",
        voice_settings: { stability, similarity_boost, style, speed, use_speaker_boost: true },
      }),
    });
    if (!r.ok) throw new Error(`scene ${i + 1} (${scene.id}): ElevenLabs answered ${r.status} ${await r.text()}`);
    const file = path.join(OUT, `${String(i + 1).padStart(2, "0")}.mp3`);
    fs.writeFileSync(file, Buffer.from(await r.arrayBuffer()));
    console.log(file, "←", scene.say);
  }
  console.log("Now run: node video/render.js --site");
})().catch((e) => { console.error(e.message); process.exit(1); });
