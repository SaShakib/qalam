// The download button: finds this visitor's system and processor, and the right file in the
// latest GitHub release.
(() => {
  const { repoUrl, gh, esc } = window.Qalam;

  // Which system and processor is this? Chrome and Edge say exactly (userAgentData);
  // others are read from the user agent, and Macs from the graphics chip.
  async function detect() {
    const ua = navigator.userAgent;
    let os = "other", arch = "", sure = false;
    const uad = navigator.userAgentData;
    if (uad && uad.platform) {
      const p = uad.platform.toLowerCase();
      os = p.includes("mac") ? "mac" : p.includes("win") ? "windows" : /android|ios|chrome ?os/.test(p) ? "mobile" : p.includes("linux") ? "linux" : "other";
      try {
        const v = await uad.getHighEntropyValues(["architecture", "bitness"]);
        if (v.architecture) { arch = v.architecture === "arm" ? "arm64" : "x64"; sure = true; }
      } catch {}
    } else if (/iPhone|iPad|iPod|Android/i.test(ua) || (/Macintosh/.test(ua) && navigator.maxTouchPoints > 1)) {
      os = "mobile";
    } else if (/Mac OS X|Macintosh/.test(ua)) {
      os = "mac";
    } else if (/Windows/.test(ua)) {
      os = "windows";
      if (/ARM64|aarch64/i.test(ua)) { arch = "arm64"; sure = true; }
    } else if (/Linux|X11/.test(ua)) {
      os = "linux";
    }
    if (os === "mac" && !arch) {
      // Safari and Firefox always report "Intel"; the graphics chip tells the truth.
      try {
        const gl = document.createElement("canvas").getContext("webgl");
        const info = gl && gl.getExtension("WEBGL_debug_renderer_info");
        const gpu = info ? gl.getParameter(info.UNMASKED_RENDERER_WEBGL) : "";
        if (/Apple (M\d|GPU)/i.test(gpu)) { arch = "arm64"; sure = true; }
        else if (/Intel|AMD|Radeon|NVIDIA/i.test(gpu)) { arch = "x64"; sure = true; }
      } catch {}
    }
    if (os === "windows" && !arch) arch = "x64";
    return { os, arch, sure };
  }

  const label = { mac: "Mac", windows: "Windows" };
  const archLabel = (d) => d.os === "mac"
    ? (d.arch === "arm64" ? "Apple Silicon (M1 or later)" : d.arch === "x64" ? "Intel" : "")
    : (d.arch === "arm64" ? "ARM64 (Snapdragon, Surface Pro X)" : "64-bit (Intel or AMD)");

  // The file for this system in a release, best first.
  function pick(assets, d) {
    const by = (re) => assets.find(a => re.test(a.name));
    if (d.os === "mac") return by(/^Qalam-mac\.pkg$/) || by(/^Qalam-[\d.]+\.pkg$/) || by(/\.pkg$/) || by(/^Qalam-mac\.zip$/);
    if (d.os === "windows") {
      const a = d.arch === "arm64" ? "arm64" : "x64";
      return by(new RegExp(`^Qalam-Setup-${a}\\.exe$`)) || by(new RegExp(`^Qalam-Windows-${a}\\.exe$`));
    }
    return null;
  }

  const size = (b) => b > 1e6 ? (b / 1e6).toFixed(1) + " MB" : Math.round(b / 1e3) + " KB";

  // What each file is, for the "other downloads" list.
  function describe(name) {
    if (/^Qalam-Setup-x64\.exe$/.test(name)) return "Windows installer, 64-bit (Intel or AMD)";
    if (/^Qalam-Setup-arm64\.exe$/.test(name)) return "Windows installer, ARM64";
    if (/^Qalam-Windows-x64\.exe$/.test(name)) return "Windows, single file (no installer), 64-bit";
    if (/^Qalam-Windows-arm64\.exe$/.test(name)) return "Windows, single file (no installer), ARM64";
    if (/\.pkg$/.test(name)) return "Mac installer";
    if (/^Qalam-mac\.zip$/.test(name)) return "Mac zip (for Homebrew and install.sh)";
    if (/SHA256/.test(name)) return "Checksums";
    return "";
  }

  async function init() {
    const btn = document.getElementById("dl-main");
    if (!btn) return;
    const title = btn.querySelector(".t"), sub = btn.querySelector(".sub");
    const detected = document.getElementById("dl-detected");
    const others = document.getElementById("dl-others");
    const note = document.getElementById("dl-note");
    const d = await detect();

    if (d.os === "mac" || d.os === "windows") {
      detected.textContent = `Detected: ${label[d.os]}${archLabel(d) ? " · " + archLabel(d) : ""}`;
      title.textContent = `Download for ${label[d.os]}`;
      sub.textContent = d.os === "mac" ? "macOS 13 Ventura or later · Apple Silicon and Intel" : "Windows 10 and 11 · " + archLabel(d);
    } else {
      detected.textContent = d.os === "mobile" ? "Detected: phone or tablet" : "Detected: " + (d.os === "linux" ? "Linux" : "another system");
      title.textContent = "Try it in your browser";
      sub.textContent = "Qalam installs on Mac and Windows";
      btn.href = "#demo";
      note.innerHTML = `The keyboard is for <a href="docs.html?p=install-mac">Mac</a> and <a href="docs.html?p=install-windows">Windows</a>. On this device, use the typing pad below and copy the Arabic.`;
    }

    let rel = null;
    try { rel = await gh("/releases/latest"); } catch {}
    if (!rel) {
      if (d.os === "mac" || d.os === "windows") btn.href = `${repoUrl}/releases/latest`;
      others.innerHTML = `<a href="${repoUrl}/releases/latest">All downloads on GitHub</a>`;
      return;
    }
    const version = rel.tag_name.replace(/^v/, "");
    const asset = pick(rel.assets, d);
    // Mac builds are universal from 1.3.0; older ones run on Apple Silicon only.
    const [major, minor] = version.split(".").map(Number);
    const universal = major > 1 || (major === 1 && minor >= 3);
    if (d.os === "mac" && !universal) {
      sub.textContent = "macOS 13 Ventura or later · Apple Silicon";
      if (d.arch === "x64") note.textContent = "This version needs an Apple Silicon Mac. Intel Macs are supported from version 1.3.0.";
    }
    if (asset) {
      btn.href = asset.browser_download_url;
      sub.textContent += ` · version ${version} · ${size(asset.size)}`;
      if (d.os === "windows" && !d.sure) {
        const other = pick(rel.assets, { os: "windows", arch: "arm64" });
        if (other) note.innerHTML = `On an ARM PC (Snapdragon, Surface Pro X)? <a href="${esc(other.browser_download_url)}">Get the ARM64 version</a>.`;
      }
    } else if (d.os === "mac" || d.os === "windows") {
      btn.href = `${repoUrl}/releases/latest`;
    }
    const files = rel.assets.filter(a => describe(a.name));
    others.innerHTML = `<details><summary>Other downloads (version ${esc(version)})</summary><ul>` +
      files.map(a => `<li><a href="${esc(a.browser_download_url)}">${esc(a.name)}</a> <span class="muted">· ${esc(describe(a.name))} · ${size(a.size)}</span></li>`).join("") +
      `<li><a href="${repoUrl}/releases">Older versions</a></li></ul></details>`;
  }

  window.QalamDownload = { detect, pick };
  init();
})();
