cask "qalam" do
  version "1.0.0"
  sha256 "7709d8c7097a01170c6c9db0538fc0328829ae403b227a9d1d037b3c3ed3b817"

  url "https://github.com/SaShakib/qalam/releases/download/v#{version}/Qalam-mac.zip"
  name "Qalam"
  desc "Phonetic Arabic keyboard with full harakat and Qur'anic marks"
  homepage "https://github.com/SaShakib/qalam"

  depends_on macos: ">= :ventura"

  app "Qalam.app"
  input_method "QalamInput.app"

  postflight do
    # The release is not notarized: clear the download quarantine so macOS will load it.
    system_command "/usr/bin/xattr",
                   args:         ["-dr", "com.apple.quarantine",
                                  "#{appdir}/Qalam.app",
                                  "#{Dir.home}/Library/Input Methods/QalamInput.app"],
                   must_succeed: false
    system_command "/usr/bin/killall", args: ["QalamInput"], must_succeed: false
  end

  uninstall quit: "com.qalam.app"

  zap trash: [
    "~/Library/Preferences/com.qalam.app.plist",
    "~/Library/Preferences/com.qalam.shared.plist",
  ]

  caveats <<~EOS
    Open Qalam once: it adds the keyboard and runs a short 20-word practice.
    Switch keyboards with the Globe key or Control-Space.
  EOS
end
