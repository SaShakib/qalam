// Renders video/qalam-demo.html into an MP4 with voice and subtitles.
//
//   make wasm                     (once: the page uses the real engine)
//   node video/elevenlabs.js      (the voice: video/voice/01.mp3 … one per scene in script.json)
//   node video/render.js          → video/out/
//   node video/render.js --site   … and copy the video to website/media/ for the website
//
// Each scene lasts as long as its voice clip, so the voice and pictures always match. Without clips the
// scene lengths are estimated and the video is silent. Output in video/out/:
//   qalam-demo.mp4        subtitles drawn in, with the voice if there is one (post this one)
//   qalam-demo-clean.mp4  no subtitles, same voice (for YouTube with the .srt)
//   qalam-demo.srt / .vtt the subtitles; poster.png a thumbnail
//
// Needs Playwright (npm i -D playwright) and ffmpeg with libx264.
const http = require("http");
const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");
const { chromium } = require("playwright");

const ROOT = path.resolve(__dirname, "..");
const OUT = path.join(__dirname, "out");
const FPS = Number(process.env.FPS || 30);
const TYPES = { ".html": "text/html", ".js": "text/javascript", ".json": "application/json", ".wasm": "application/wasm", ".ttf": "font/ttf" };

function serve() {
  const server = http.createServer((req, res) => {
    const file = path.join(ROOT, decodeURIComponent(new URL(req.url, "http://x").pathname));
    if (!file.startsWith(ROOT) || !fs.existsSync(file) || fs.statSync(file).isDirectory()) { res.writeHead(404); return res.end(); }
    res.writeHead(200, { "Content-Type": TYPES[path.extname(file)] || "application/octet-stream" });
    fs.createReadStream(file).pipe(res);
  });
  return new Promise((r) => server.listen(0, "127.0.0.1", () => r(server)));
}

const stamp = (t) => {
  const ms = Math.round(t * 1000), h = Math.floor(ms / 3600000), m = Math.floor(ms / 60000) % 60, s = Math.floor(ms / 1000) % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")},${String(ms % 1000).padStart(3, "0")}`;
};
const seconds = (file) => Number(execFileSync("ffprobe", ["-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", file]).toString());
const ffmpeg = (args) => execFileSync("ffmpeg", ["-y", "-loglevel", "error", ...args], { stdio: "inherit" });

(async () => {
  if (!fs.existsSync(path.join(ROOT, "website/demo/qalam.wasm"))) throw new Error("Run `make wasm` first.");
  const script = JSON.parse(fs.readFileSync(path.join(__dirname, "script.json"), "utf8"));
  const clipDir = path.join(__dirname, "voice");
  const clips = fs.existsSync(clipDir) ? fs.readdirSync(clipDir).filter((f) => /^\d+\.(mp3|wav|m4a)$/.test(f)).sort() : [];
  let durations = null;
  if (clips.length) {
    if (clips.length !== script.scenes.length) throw new Error(`video/voice/ has ${clips.length} clips but script.json has ${script.scenes.length} scenes: run node video/elevenlabs.js`);
    durations = clips.map((f) => seconds(path.join(clipDir, f)));
  } else {
    console.log("No voice clips in video/voice/: the video will be silent, with estimated timing.");
  }

  fs.mkdirSync(OUT, { recursive: true });
  const frames = path.join(OUT, "frames"), clean = path.join(OUT, "frames-clean");
  for (const d of [frames, clean]) { fs.rmSync(d, { recursive: true, force: true }); fs.mkdirSync(d, { recursive: true }); }

  const server = await serve();
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1920, height: 1080 } });
  page.on("pageerror", (e) => console.error("page error:", e.message));
  await page.goto(`http://127.0.0.1:${server.address().port}/video/qalam-demo.html?capture`);
  await page.waitForFunction(() => window.videoReady, null, { timeout: 30000 });
  await page.evaluate(() => window.videoReady);
  const timeline = await page.evaluate((d) => d ? window.setVoiceDurations(d) : window.setVoiceDurations(Array(0)), durations);
  const { duration, scenes, subtitles } = timeline;
  console.log(`${scenes.length} scenes, ${duration.toFixed(1)} s`);

  // Each frame twice: with the page's subtitles, and without.
  const total = Math.round(duration * FPS);
  for (let f = 0; f < total; f++) {
    const name = `${String(f).padStart(5, "0")}.png`;
    await page.evaluate((t) => { window.renderAt(t); document.getElementById("sub").style.visibility = ""; }, f / FPS);
    await page.screenshot({ path: path.join(frames, name) });
    await page.evaluate(() => { document.getElementById("sub").style.visibility = "hidden"; });
    await page.screenshot({ path: path.join(clean, name) });
    if (f % (FPS * 5) === 0) process.stdout.write(`\r${Math.round(f / FPS)} s / ${Math.round(duration)} s`);
  }
  const intro = scenes.find((s) => s.id === "intro");
  await page.evaluate((t) => { window.renderAt(t); document.getElementById("sub").style.visibility = "hidden"; }, intro ? Math.min(2.2, intro.len - 0.5) : 1);
  await page.screenshot({ path: path.join(OUT, "poster.png") });
  await browser.close();
  server.close();
  console.log("\nframes done");

  const srt = subtitles.map((s, i) => `${i + 1}\n${stamp(s.start)} --> ${stamp(s.end)}\n${s.text}\n`).join("\n");
  fs.writeFileSync(path.join(OUT, "qalam-demo.srt"), srt);
  fs.writeFileSync(path.join(OUT, "qalam-demo.vtt"), "WEBVTT\n\n" + srt.replace(/(\d\d:\d\d:\d\d),(\d\d\d)/g, "$1.$2"));

  // The voice track: each clip where its scene's narration starts.
  let voice = null;
  if (durations) {
    voice = path.join(OUT, "voice.wav");
    const inputs = clips.flatMap((f) => ["-i", path.join(clipDir, f)]);
    const graph = clips.map((_, i) => `[${i}:a]aresample=48000,adelay=${Math.round(scenes[i].voiceStart * 1000)}:all=1[a${i}]`).join(";") +
      ";" + clips.map((_, i) => `[a${i}]`).join("") + `amix=inputs=${clips.length}:normalize=0,apad[out]`;
    ffmpeg([...inputs, "-filter_complex", graph, "-map", "[out]", "-t", duration.toFixed(3), voice]);
    console.log("voice: placed", clips.length, "clips");
  }

  const enc = ["-c:v", "libx264", "-preset", "slow", "-crf", "20", "-pix_fmt", "yuv420p"];
  const encode = (dir, out) => ffmpeg(["-framerate", String(FPS), "-i", path.join(dir, "%05d.png"), ...(voice ? ["-i", voice] : []), ...enc,
    ...(voice ? ["-c:a", "aac", "-b:a", "160k", "-shortest"] : []), "-movflags", "+faststart", path.join(OUT, out)]);
  encode(frames, "qalam-demo.mp4");
  encode(clean, "qalam-demo-clean.mp4");
  for (const d of [frames, clean]) fs.rmSync(d, { recursive: true, force: true });
  for (const old of ["qalam-demo-voice.mp4", "voice-synced.wav"]) fs.rmSync(path.join(OUT, old), { force: true });

  if (process.argv.includes("--site")) {
    // The site shows its own captions, so it gets the clean video.
    const media = path.join(ROOT, "website", "media");
    fs.mkdirSync(media, { recursive: true });
    fs.copyFileSync(path.join(OUT, "qalam-demo-clean.mp4"), path.join(media, "qalam-demo.mp4"));
    ffmpeg(["-i", path.join(OUT, "qalam-demo-clean.mp4"), "-c:v", "libvpx-vp9", "-crf", "34", "-b:v", "0", "-row-mt", "1",
      ...(voice ? ["-c:a", "libopus", "-b:a", "96k"] : ["-an"]), path.join(media, "qalam-demo.webm")]);
    fs.copyFileSync(path.join(OUT, "poster.png"), path.join(media, "qalam-demo-poster.png"));
    fs.copyFileSync(path.join(OUT, "qalam-demo.vtt"), path.join(media, "qalam-demo.vtt"));
    console.log("copied to website/media/");
  }
  console.log("wrote", fs.readdirSync(OUT).map((f) => "video/out/" + f).join(", "));
})().catch((e) => { console.error(e); process.exit(1); });
