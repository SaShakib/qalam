cask "qalam" do
  version "1.1.2"
  sha256 "d2d5bb417e13f05d34f100243d9ff375b476c5b49d25082e8c0072ca813fe3ee"

  url "https://github.com/SaShakib/qalam/releases/download/v#{version}/Qalam-mac.zip"
  name "Qalam"
  desc "Phonetic Arabic keyboard with full harakat and Qur'anic marks"
  homepage "https://github.com/SaShakib/qalam"

  depends_on macos: :ventura

  app "Qalam.app"
  input_method "QalamInput.app"

  # The release is not notarized: clear the download quarantine so macOS will load it,
  # then restart the keyboard so a new version is used.
  postflight_steps do
    run "/bin/sh",
        args:         ["-c", 'xattr -dr com.apple.quarantine "{{appdir}}/Qalam.app" ' \
                             '"$HOME/Library/Input Methods/QalamInput.app" 2>/dev/null; true'],
        must_succeed: false
    terminate_process "QalamInput"
    terminate_process "TextInputMenuAgent"
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
