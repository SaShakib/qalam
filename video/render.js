// Renders video/qalam-demo.html into an MP4 with subtitles.
//
//   make wasm                 (once: the page uses the real engine)
//   node video/render.js      → video/out/qalam-demo.mp4 (with subtitles),
//                               qalam-demo-clean.mp4 (no subtitles), qalam-demo.srt / .vtt, poster.png
//   With a voice-over, also qalam-demo-voice.mp4: either one file, video/voice.mp3 (or .wav / .m4a),
//   or one clip per subtitle line in video/voice/01.mp3 … 07.mp3, which are lined up automatically.
//   node video/render.js --site  also copies the video, poster and captions into website/media/.
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
const TYPES = { ".html": "text/html", ".js": "text/javascript", ".wasm": "application/wasm", ".ttf": "font/ttf", ".css": "text/css" };

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

(async () => {
  if (!fs.existsSync(path.join(ROOT, "website/demo/qalam.wasm"))) throw new Error("Run `make wasm` first.");
  const frames = path.join(OUT, "frames");
  fs.rmSync(frames, { recursive: true, force: true });
  fs.mkdirSync(frames, { recursive: true });

  const server = await serve();
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1920, height: 1080 } });
  page.on("pageerror", (e) => console.error("page error:", e.message));
  await page.goto(`http://127.0.0.1:${server.address().port}/video/qalam-demo.html?capture`);
  await page.waitForFunction(() => window.videoReady && window.renderAt, null, { timeout: 30000 });
  await page.evaluate(() => window.videoReady);
  const [duration, subs] = await page.evaluate(() => [window.DURATION, window.SUBTITLES]);

  // Each frame twice: with the page's subtitles, and without (for a clean video plus the .srt).
  const clean = path.join(OUT, "frames-clean");
  fs.rmSync(clean, { recursive: true, force: true });
  fs.mkdirSync(clean, { recursive: true });
  const total = Math.round(duration * FPS);
  for (let f = 0; f < total; f++) {
    const name = `${String(f).padStart(4, "0")}.png`;
    await page.evaluate((t) => { window.renderAt(t); document.getElementById("sub").style.visibility = ""; }, f / FPS);
    await page.screenshot({ path: path.join(frames, name) });
    await page.evaluate(() => { document.getElementById("sub").style.visibility = "hidden"; });
    await page.screenshot({ path: path.join(clean, name) });
    if (f % FPS === 0) process.stdout.write(`\r${f / FPS}s / ${duration}s`);
  }
  await page.evaluate((t) => window.renderAt(t), 1.6);
  await page.screenshot({ path: path.join(OUT, "poster.png") });
  await browser.close();
  server.close();
  console.log("\nframes done");

  const srt = subs.map(([a, b, text], i) => `${i + 1}\n${stamp(a)} --> ${stamp(b)}\n${text}\n`).join("\n");
  fs.writeFileSync(path.join(OUT, "qalam-demo.srt"), srt);
  // WebVTT for the website's <video> captions.
  fs.writeFileSync(path.join(OUT, "qalam-demo.vtt"), "WEBVTT\n\n" + srt.replace(/(\d\d:\d\d:\d\d),(\d\d\d)/g, "$1.$2"));

  const enc = ["-c:v", "libx264", "-preset", "slow", "-crf", "20", "-pix_fmt", "yuv420p", "-movflags", "+faststart"];
  const encode = (dir, out) => execFileSync("ffmpeg", ["-y", "-loglevel", "error", "-framerate", String(FPS),
    "-i", path.join(dir, "%04d.png"), ...enc, path.join(OUT, out)], { stdio: "inherit" });
  encode(frames, "qalam-demo.mp4");
  encode(clean, "qalam-demo-clean.mp4");
  fs.rmSync(clean, { recursive: true, force: true });
  fs.rmSync(frames, { recursive: true, force: true });

  // The voice-over, if there is one: put it under the subtitled video (the video stays 30 s).
  // Either one file (video/voice.mp3), or one clip per subtitle line in video/voice/ (01.mp3 … 07.mp3,
  // in subtitle order): each clip starts with its subtitle, and is sped up a little if it's longer than its slot.
  let voice = ["voice.mp3", "voice.wav", "voice.m4a"].map((f) => path.join(__dirname, f)).find((f) => fs.existsSync(f));
  const clipDir = path.join(__dirname, "voice");
  const clips = fs.existsSync(clipDir) ? fs.readdirSync(clipDir).filter((f) => /^\d+\.(mp3|wav|m4a)$/.test(f)).sort() : [];
  if (!voice && clips.length) {
    if (clips.length !== subs.length) throw new Error(`video/voice/ has ${clips.length} clips but there are ${subs.length} subtitle lines`);
    const args = ["-y", "-loglevel", "error"];
    const parts = [];
    clips.forEach((f, i) => {
      const file = path.join(clipDir, f);
      args.push("-i", file);
      const len = Number(execFileSync("ffprobe", ["-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", file]).toString());
      const [start] = subs[i];
      const slot = (i + 1 < subs.length ? subs[i + 1][0] : duration) - start - 0.1;
      const tempo = Math.min(Math.max(len / slot, 1), 1.3);
      if (len / slot > 1.3) console.warn(`line ${i + 1} is ${len.toFixed(1)} s for a ${slot.toFixed(1)} s slot: shorten its text`);
      parts.push(`[${i}:a]atempo=${tempo.toFixed(3)},adelay=${Math.round(start * 1000)}:all=1[a${i}]`);
    });
    const mix = parts.join(";") + ";" + clips.map((_, i) => `[a${i}]`).join("") + `amix=inputs=${clips.length}:normalize=0,apad[out]`;
    voice = path.join(OUT, "voice-synced.wav");
    execFileSync("ffmpeg", [...args, "-filter_complex", mix, "-map", "[out]", "-t", String(duration), voice], { stdio: "inherit" });
    console.log("voice: placed", clips.length, "clips at their subtitles");
  }
  if (voice) {
    execFileSync("ffmpeg", ["-y", "-loglevel", "error", "-i", path.join(OUT, "qalam-demo.mp4"), "-i", voice,
      "-map", "0:v", "-map", "1:a", "-c:v", "copy", "-c:a", "aac", "-b:a", "160k", "-af", "apad", "-t", String(duration),
      "-movflags", "+faststart", path.join(OUT, "qalam-demo-voice.mp4")], { stdio: "inherit" });
  }
  if (process.argv.includes("--site")) {
    const media = path.join(ROOT, "website", "media");
    fs.mkdirSync(media, { recursive: true });
    // The site shows its own captions, so it gets the clean video (with the voice, if there is one).
    const src = path.join(media, "qalam-demo.mp4");
    if (voice) {
      execFileSync("ffmpeg", ["-y", "-loglevel", "error", "-i", path.join(OUT, "qalam-demo-clean.mp4"), "-i", voice,
        "-map", "0:v", "-map", "1:a", "-c:v", "copy", "-c:a", "aac", "-b:a", "160k", "-af", "apad", "-t", String(duration),
        "-movflags", "+faststart", src], { stdio: "inherit" });
    } else {
      fs.copyFileSync(path.join(OUT, "qalam-demo-clean.mp4"), src);
    }
    // WebM (VP9) too, for browsers without H.264.
    execFileSync("ffmpeg", ["-y", "-loglevel", "error", "-i", src, "-c:v", "libvpx-vp9", "-crf", "34", "-b:v", "0", "-row-mt", "1",
      ...(voice ? ["-c:a", "libopus", "-b:a", "96k"] : ["-an"]), path.join(media, "qalam-demo.webm")], { stdio: "inherit" });
    fs.copyFileSync(path.join(OUT, "poster.png"), path.join(media, "qalam-demo-poster.png"));
    fs.copyFileSync(path.join(OUT, "qalam-demo.vtt"), path.join(media, "qalam-demo.vtt"));
    console.log("copied to website/media/");
  }
  console.log("wrote", fs.readdirSync(OUT).map((f) => "video/out/" + f).join(", "));
})().catch((e) => { console.error(e); process.exit(1); });
