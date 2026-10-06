#!/bin/bash
# Qalam installer for macOS.
#
#   Install:    curl -fsSL https://raw.githubusercontent.com/SaShakib/qalam/main/install.sh | bash
#   Uninstall:  curl -fsSL https://raw.githubusercontent.com/SaShakib/qalam/main/install.sh | bash -s -- --uninstall
#   From a downloaded zip:  bash install.sh Qalam-mac.zip
#
# Installs for the current user only (no password needed):
#   ~/Applications/Qalam.app                    the guide and practice app
#   ~/Library/Input Methods/QalamInput.app      the keyboard
set -euo pipefail

REPO="SaShakib/qalam"
ZIP_URL="https://github.com/$REPO/releases/latest/download/Qalam-mac.zip"
IME="$HOME/Library/Input Methods/QalamInput.app"
APP="$HOME/Applications/Qalam.app"

if [ "$(uname)" != "Darwin" ]; then
  echo "This installer is for macOS. On Windows, download Qalam-Windows-x64.exe from:"
  echo "  https://github.com/$REPO/releases/latest"
  exit 1
fi

if [ "${1:-}" = "--uninstall" ]; then
  killall QalamInput Qalam 2>/dev/null || true
  rm -rf "$IME" "$APP"
  echo "Qalam removed. If it is still listed, remove it in System Settings → Keyboard → Input Sources."
  exit 0
fi

major=$(sw_vers -productVersion | cut -d. -f1)
if [ "$major" -lt 13 ]; then
  echo "Qalam needs macOS 13 (Ventura) or later. This Mac has $(sw_vers -productVersion)."
  exit 1
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

if [ -n "${1:-}" ] && [ -f "$1" ]; then
  cp "$1" "$tmp/qalam.zip"
else
  echo "Downloading Qalam…"
  curl -fL --progress-bar "$ZIP_URL" -o "$tmp/qalam.zip"
fi

ditto -x -k "$tmp/qalam.zip" "$tmp/x"

killall QalamInput Qalam 2>/dev/null || true
mkdir -p "$HOME/Library/Input Methods" "$HOME/Applications"
rm -rf "$IME" "$APP"
ditto "$tmp/x/QalamInput.app" "$IME"
ditto "$tmp/x/Qalam.app" "$APP"
xattr -dr com.apple.quarantine "$IME" "$APP" 2>/dev/null || true

if [ -d "/Applications/Qalam.app" ] || [ -d "/Library/Input Methods/QalamInput.app" ]; then
  echo "Note: an older Qalam from the .pkg installer is also present in /Applications."
  echo "      Remove it to avoid two copies: sudo rm -rf /Applications/Qalam.app '/Library/Input Methods/QalamInput.app'"
fi

echo "Qalam is installed. Opening it now: it will help you add the keyboard and practise."
sleep 1
open "$APP" 2>/dev/null || { sleep 2; open "$APP"; }
