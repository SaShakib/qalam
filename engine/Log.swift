import Foundation

/// A short log of what Qalam did (~/Library/Logs/Qalam.log), for "it didn't insert" reports.
/// Kept small: started afresh once it passes 64 KB.
public enum QalamLog {
    static let url = FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Library/Logs/Qalam.log")

    public static func note(_ message: String, from: String = ProcessInfo.processInfo.processName) {
        let line = "\(ISO8601DateFormatter().string(from: Date())) \(from): \(message)\n"
        let fm = FileManager.default
        if let size = (try? fm.attributesOfItem(atPath: url.path))?[.size] as? Int, size > 65_536 {
            try? fm.removeItem(at: url)
        }
        if let h = try? FileHandle(forWritingTo: url) {
            h.seekToEndOfFile()
            h.write(Data(line.utf8))
            try? h.close()
        } else {
            try? Data(line.utf8).write(to: url)
        }
    }
}
