// Draws the app icon (AppIcon.iconset) and the menu-bar icon (icon.tiff).
// Usage: swift tools/make_icons.swift <output dir>
import AppKit

let outDir = URL(fileURLWithPath: CommandLine.arguments.count > 1 ? CommandLine.arguments[1] : "build/icons")
try? FileManager.default.createDirectory(at: outDir, withIntermediateDirectories: true)

func bitmap(_ px: Int, _ draw: (CGFloat) -> Void) -> NSBitmapImageRep {
    let rep = NSBitmapImageRep(bitmapDataPlanes: nil, pixelsWide: px, pixelsHigh: px, bitsPerSample: 8,
                               samplesPerPixel: 4, hasAlpha: true, isPlanar: false, colorSpaceName: .deviceRGB,
                               bytesPerRow: 0, bitsPerPixel: 0)!
    NSGraphicsContext.saveGraphicsState()
    NSGraphicsContext.current = NSGraphicsContext(bitmapImageRep: rep)
    draw(CGFloat(px))
    NSGraphicsContext.restoreGraphicsState()
    return rep
}

/// Draws `text` centred on its ink, not on its line box (Arabic glyphs sit oddly in line boxes).
func drawCentered(_ text: String, size: CGFloat, color: NSColor, in side: CGFloat, nudgeY: CGFloat = 0) {
    let font = NSFont(name: "GeezaPro-Bold", size: size) ?? NSFont.boldSystemFont(ofSize: size)
    let s = NSAttributedString(string: text, attributes: [.font: font, .foregroundColor: color])
    let ctx = NSGraphicsContext.current!.cgContext
    let line = CTLineCreateWithAttributedString(s)
    let ink = CTLineGetImageBounds(line, ctx)
    ctx.textPosition = CGPoint(x: side / 2 - ink.midX, y: side / 2 - ink.midY + nudgeY)
    CTLineDraw(line, ctx)
}

func appIcon(_ px: Int) -> NSBitmapImageRep {
    bitmap(px) { side in
        let inset = side * 0.1
        let rect = CGRect(x: inset, y: inset, width: side - 2 * inset, height: side - 2 * inset)
        let path = NSBezierPath(roundedRect: rect, xRadius: side * 0.18, yRadius: side * 0.18)
        NSGradient(starting: NSColor(calibratedRed: 0.05, green: 0.47, blue: 0.42, alpha: 1),
                   ending: NSColor(calibratedRed: 0.02, green: 0.30, blue: 0.33, alpha: 1))!.draw(in: path, angle: -90)
        drawCentered("قَ", size: side * 0.5, color: .white, in: side)
    }
}

func png(_ rep: NSBitmapImageRep, _ name: String, in dir: URL) {
    try! rep.representation(using: .png, properties: [:])!.write(to: dir.appendingPathComponent(name))
}

// App icon set
let iconset = outDir.appendingPathComponent("AppIcon.iconset")
try? FileManager.default.createDirectory(at: iconset, withIntermediateDirectories: true)
for base in [16, 32, 128, 256, 512] {
    png(appIcon(base), "icon_\(base)x\(base).png", in: iconset)
    png(appIcon(base * 2), "icon_\(base)x\(base)@2x.png", in: iconset)
}
png(appIcon(512), "preview.png", in: outDir)

// Menu-bar icon: black ق on transparent (macOS tints it as a template).
func menuIcon(_ px: Int) -> NSBitmapImageRep {
    bitmap(px) { side in
        drawCentered("ق", size: side * 0.95, color: .black, in: side)
    }
}
let small = menuIcon(16), large = menuIcon(32)
large.size = NSSize(width: 16, height: 16)
let image = NSImage(size: NSSize(width: 16, height: 16))
image.addRepresentation(small)
image.addRepresentation(large)
try! image.tiffRepresentation!.write(to: outDir.appendingPathComponent("icon.tiff"))
png(menuIcon(64), "menu-preview.png", in: outDir)
print("icons written to \(outDir.path)")

// Windows icons (.ico with PNG images): teal = Arabic on, grey = off.
func trayIcon(_ px: Int, on: Bool) -> NSBitmapImageRep {
    bitmap(px) { side in
        let rect = CGRect(x: 0, y: 0, width: side, height: side).insetBy(dx: side * 0.04, dy: side * 0.04)
        let path = NSBezierPath(roundedRect: rect, xRadius: side * 0.22, yRadius: side * 0.22)
        (on ? NSColor(calibratedRed: 0.04, green: 0.42, blue: 0.40, alpha: 1)
            : NSColor(calibratedWhite: 0.55, alpha: 1)).setFill()
        path.fill()
        drawCentered("ق", size: side * 0.78, color: .white, in: side)
    }
}

func writeICO(_ name: String, on: Bool) {
    let sizes = [16, 20, 24, 32, 40, 48, 64, 256]
    let pngs = sizes.map { trayIcon($0, on: on).representation(using: .png, properties: [:])! }
    var data = Data()
    func u16(_ v: Int) { data.append(UInt8(v & 0xFF)); data.append(UInt8((v >> 8) & 0xFF)) }
    func u32(_ v: Int) { u16(v & 0xFFFF); u16((v >> 16) & 0xFFFF) }
    u16(0); u16(1); u16(sizes.count)
    var offset = 6 + 16 * sizes.count
    for (i, s) in sizes.enumerated() {
        data.append(UInt8(s >= 256 ? 0 : s)); data.append(UInt8(s >= 256 ? 0 : s))
        data.append(0); data.append(0)
        u16(1); u16(32)
        u32(pngs[i].count); u32(offset)
        offset += pngs[i].count
    }
    for p in pngs { data.append(p) }
    try! data.write(to: outDir.appendingPathComponent(name))
}
writeICO("qalam-on.ico", on: true)
writeICO("qalam-off.ico", on: false)
png(trayIcon(64, on: true), "tray-preview.png", in: outDir)
print("windows icons written")
