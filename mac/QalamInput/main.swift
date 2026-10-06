import Cocoa
import InputMethodKit

// macOS starts this process when Qalam is chosen in the input menu.
let connectionName = Bundle.main.infoDictionary?["InputMethodConnectionName"] as? String
    ?? "com.qalam.inputmethod.Qalam_Connection"
let server = IMKServer(name: connectionName, bundleIdentifier: Bundle.main.bundleIdentifier)
_ = server
NSApplication.shared.run()
