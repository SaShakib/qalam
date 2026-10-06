cask "qalam" do
  version "1.1.1"
  sha256 "a4fdb03153fbf4c57fba575885c810b82af6be397628c989e5f302ae99f644e0"

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
