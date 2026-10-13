// Renders video/qalam-demo.html into an MP4 with subtitles.
//
//   make wasm                 (once: the page uses the real engine)
//   node video/render.js      → video/out/qalam-demo.mp4 (with subtitles),
//                               qalam-demo-clean.mp4 (no subtitles), qalam-demo.srt / .vtt, poster.png
//   With a voice-over in video/voice.mp3 (or .wav / .m4a), also qalam-demo-voice.mp4.
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
  const voice = ["voice.mp3", "voice.wav", "voice.m4a"].map((f) => path.join(__dirname, f)).find((f) => fs.existsSync(f));
  if (voice) {
    execFileSync("ffmpeg", ["-y", "-loglevel", "error", "-i", path.join(OUT, "qalam-demo.mp4"), "-i", voice,
      "-map", "0:v", "-map", "1:a", "-c:v", "copy", "-c:a", "aac", "-b:a", "160k", "-af", "apad", "-t", String(duration),
      "-movflags", "+faststart", path.join(OUT, "qalam-demo-voice.mp4")], { stdio: "inherit" });
  }
  if (process.argv.includes("--site")) {
    const media = path.join(ROOT, "website", "media");
    fs.mkdirSync(media, { recursive: true });
    const src = path.join(OUT, voice ? "qalam-demo-voice.mp4" : "qalam-demo-clean.mp4");
    fs.copyFileSync(src, path.join(media, "qalam-demo.mp4"));
    // WebM (VP9) too, for browsers without H.264.
    execFileSync("ffmpeg", ["-y", "-loglevel", "error", "-i", src, "-c:v", "libvpx-vp9", "-crf", "34", "-b:v", "0", "-row-mt", "1",
      ...(voice ? ["-c:a", "libopus", "-b:a", "96k"] : ["-an"]), path.join(media, "qalam-demo.webm")], { stdio: "inherit" });
    fs.copyFileSync(path.join(OUT, "poster.png"), path.join(media, "qalam-demo-poster.png"));
    fs.copyFileSync(path.join(OUT, "qalam-demo.vtt"), path.join(media, "qalam-demo.vtt"));
    console.log("copied to website/media/");
  }
  console.log("wrote", fs.readdirSync(OUT).map((f) => "video/out/" + f).join(", "));
})().catch((e) => { console.error(e); process.exit(1); });
