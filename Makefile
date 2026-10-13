# Qalam: phonetic Arabic keyboard for macOS and Windows
#   make test        run the transliteration tests (Swift and Go engines)
#   make install     build and install on this Mac (for development)
#   make release     build everything for a GitHub release into dist/
#   make uninstall   remove the development install
#   make site        build the website (with the live typing demo) into build/site
#
# Trusted Mac builds (no "Open Anyway" for users, ever): see docs/developers/signing.md.
#   make release SIGN_ID="Developer ID Application: …" INSTALLER_ID="Developer ID Installer: …" NOTARY_PROFILE=qalam

VERSION  := 1.3.0
# Builds run at low priority on 2 cores so the Mac stays responsive.
SWIFT    := nice -n 15 swift build -c release -j 2
GO       := nice -n 15 go
RELEASE  := .build/release
BUILD    := build
DIST     := dist
BIN      := $(BUILD)/bin
IME_APP  := $(BUILD)/QalamInput.app
APP      := $(BUILD)/Qalam.app
IME_DEST := $(HOME)/Library/Input Methods
APP_DEST := $(HOME)/Applications

# The Mac apps are universal (Apple Silicon and Intel). `make ARCHS=arm64` builds one for speed.
ARCHS    ?= arm64 x86_64
MACOS_MIN := 13.0

# Signing. "-" is ad-hoc (works on the Mac that built it; downloads need "Open Anyway").
# A Developer ID identity adds the hardened runtime and a secure timestamp, as notarization requires.
SIGN_ID        ?= -
INSTALLER_ID   ?=
NOTARY_PROFILE ?=
ifeq ($(SIGN_ID),-)
SIGN_FLAGS :=
else
SIGN_FLAGS := --options runtime --timestamp
endif

.PHONY: all build test icons bundle install uninstall clean guide windows windows-installer mac-release release site wasm

all: bundle

# One build per processor, joined with lipo into $(BIN).
build:
	mkdir -p $(BIN)
	for arch in $(ARCHS); do $(SWIFT) --triple $$arch-apple-macosx$(MACOS_MIN); done
	for p in QalamInput QalamApp; do \
	  lipo -create $$(for arch in $(ARCHS); do echo "$$($(SWIFT) --triple $$arch-apple-macosx$(MACOS_MIN) --show-bin-path)/$$p"; done) -output $(BIN)/$$p; \
	done
	lipo -info $(BIN)/QalamInput $(BIN)/QalamApp

test:
	$(SWIFT) --product qalam
	$(RELEASE)/qalam --test tests/cases.tsv
	cd windows && $(GO) test ./...

icons: $(BUILD)/icons/AppIcon.icns

$(BUILD)/icons/AppIcon.icns: tools/make_icons.swift
	nice -n 15 swift tools/make_icons.swift $(BUILD)/icons
	iconutil -c icns $(BUILD)/icons/AppIcon.iconset -o $(BUILD)/icons/AppIcon.icns

bundle: build icons
	rm -rf "$(IME_APP)" "$(APP)"
	mkdir -p "$(IME_APP)/Contents/MacOS" "$(IME_APP)/Contents/Resources/en.lproj"
	cp $(BIN)/QalamInput "$(IME_APP)/Contents/MacOS/QalamInput"
	cp mac/bundle/QalamInput-Info.plist "$(IME_APP)/Contents/Info.plist"
	cp mac/bundle/InfoPlist.strings "$(IME_APP)/Contents/Resources/en.lproj/InfoPlist.strings"
	cp $(BUILD)/icons/icon.tiff "$(IME_APP)/Contents/Resources/icon.tiff"
	mkdir -p "$(APP)/Contents/MacOS" "$(APP)/Contents/Resources/Fonts"
	cp $(BIN)/QalamApp "$(APP)/Contents/MacOS/Qalam"
	cp mac/bundle/Qalam-Info.plist "$(APP)/Contents/Info.plist"
	cp $(BUILD)/icons/AppIcon.icns "$(APP)/Contents/Resources/AppIcon.icns"
	cp fonts/*.ttf "$(APP)/Contents/Resources/Fonts/"
	cp fonts/*.txt "$(APP)/Contents/Resources/Fonts/"
	for p in "$(IME_APP)" "$(APP)"; do \
	  /usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $(VERSION)" "$$p/Contents/Info.plist"; \
	  /usr/libexec/PlistBuddy -c "Set :CFBundleVersion $(VERSION)" "$$p/Contents/Info.plist"; \
	done
	codesign --force $(SIGN_FLAGS) --sign "$(SIGN_ID)" "$(IME_APP)"
	codesign --force $(SIGN_FLAGS) --entitlements mac/bundle/Qalam.entitlements --sign "$(SIGN_ID)" "$(APP)"
	codesign --verify --strict --verbose=1 "$(IME_APP)" "$(APP)"

install: test bundle
	mkdir -p "$(IME_DEST)" "$(APP_DEST)"
	-killall QalamInput Qalam 2>/dev/null
	rm -rf "$(IME_DEST)/QalamInput.app" "$(APP_DEST)/Qalam.app"
	cp -R "$(IME_APP)" "$(IME_DEST)/"
	cp -R "$(APP)" "$(APP_DEST)/"
	@echo "Installed: $(IME_DEST)/QalamInput.app and $(APP_DEST)/Qalam.app"
	open "$(APP_DEST)/Qalam.app"

uninstall:
	-killall QalamInput Qalam 2>/dev/null
	rm -rf "$(IME_DEST)/QalamInput.app" "$(APP_DEST)/Qalam.app"
	@echo "Removed. Also remove Qalam from System Settings → Keyboard → Input Sources if it is still listed."

# ---------- Windows ----------

guide: icons
	$(SWIFT) --product qalam
	mkdir -p windows/assets
	$(RELEASE)/qalam --guide windows/assets/guide.html windows
	cp $(BUILD)/icons/qalam-on.ico $(BUILD)/icons/qalam-off.ico windows/assets/
	# The fonts the Windows app carries and installs: SIL Scheherazade New, and KFGQPC Hafs when it's in fonts/.
	mkdir -p windows/assets/fonts
	cp fonts/ScheherazadeNew.ttf fonts/OFL-ScheherazadeNew.txt windows/assets/fonts/
	for f in fonts/*KFGQPC* fonts/*Hafs* fonts/*HAFS*; do [ -f "$$f" ] && cp "$$f" windows/assets/fonts/ || true; done

windows: guide
	mkdir -p $(DIST)
	cd windows && for arch in amd64 arm64; do \
	  $(GO) run github.com/akavel/rsrc@v0.10.2 -arch $$arch -ico assets/qalam-on.ico -manifest assets/qalam.manifest -o rsrc_windows_$$arch.syso; \
	done
	cd windows && GOOS=windows GOARCH=amd64 $(GO) build -p 2 -trimpath -ldflags "-H windowsgui -s -w -X main.version=$(VERSION)" -o ../$(DIST)/Qalam-Windows-x64.exe .
	cd windows && GOOS=windows GOARCH=arm64 $(GO) build -p 2 -trimpath -ldflags "-H windowsgui -s -w -X main.version=$(VERSION)" -o ../$(DIST)/Qalam-Windows-arm64.exe .

# Qalam-Setup-x64.exe / -arm64.exe. Needs Inno Setup 6 (iscc), so it runs on Windows (CI does it).
windows-installer:
	cd windows/installer && iscc /DAppVersion=$(VERSION) /DArch=x64 /DExe=..\..\$(DIST)\Qalam-Windows-x64.exe qalam.iss
	cd windows/installer && iscc /DAppVersion=$(VERSION) /DArch=arm64 /DExe=..\..\$(DIST)\Qalam-Windows-arm64.exe qalam.iss

# ---------- Mac release: zip (Homebrew + install.sh) and .pkg (double-click) ----------

mac-release: test bundle
	mkdir -p $(DIST)
	rm -rf $(BUILD)/zip $(BUILD)/pkgroot $(DIST)/Qalam-mac.zip $(DIST)/Qalam-*.pkg
	mkdir -p $(BUILD)/zip
	ditto --norsrc --noextattr "$(APP)" "$(BUILD)/zip/Qalam.app"
	ditto --norsrc --noextattr "$(IME_APP)" "$(BUILD)/zip/QalamInput.app"
	cd $(BUILD)/zip && ditto -c -k --norsrc --noextattr . ../../$(DIST)/Qalam-mac.zip
ifneq ($(NOTARY_PROFILE),)
	xcrun notarytool submit $(DIST)/Qalam-mac.zip --keychain-profile "$(NOTARY_PROFILE)" --wait
	xcrun stapler staple "$(BUILD)/zip/Qalam.app"
	xcrun stapler staple "$(BUILD)/zip/QalamInput.app"
	rm -f $(DIST)/Qalam-mac.zip
	cd $(BUILD)/zip && ditto -c -k --norsrc --noextattr . ../../$(DIST)/Qalam-mac.zip
endif
	mkdir -p "$(BUILD)/pkgroot/Applications" "$(BUILD)/pkgroot/Library/Input Methods"
	ditto --norsrc --noextattr "$(BUILD)/zip/Qalam.app" "$(BUILD)/pkgroot/Applications/Qalam.app"
	ditto --norsrc --noextattr "$(BUILD)/zip/QalamInput.app" "$(BUILD)/pkgroot/Library/Input Methods/QalamInput.app"
	pkgbuild --analyze --root $(BUILD)/pkgroot $(BUILD)/components.plist
	for i in 0 1; do /usr/libexec/PlistBuddy -c "Set :$$i:BundleIsRelocatable false" $(BUILD)/components.plist; done
	xattr -cr $(BUILD)/pkgroot
	pkgbuild --root $(BUILD)/pkgroot --component-plist $(BUILD)/components.plist --scripts mac/pkg/scripts \
	  --identifier com.qalam.pkg --version $(VERSION) --install-location / $(BUILD)/Qalam-unsigned.pkg
ifneq ($(INSTALLER_ID),)
	productsign --timestamp --sign "$(INSTALLER_ID)" $(BUILD)/Qalam-unsigned.pkg $(DIST)/Qalam-$(VERSION).pkg
else
	cp $(BUILD)/Qalam-unsigned.pkg $(DIST)/Qalam-$(VERSION).pkg
endif
ifneq ($(NOTARY_PROFILE),)
	xcrun notarytool submit $(DIST)/Qalam-$(VERSION).pkg --keychain-profile "$(NOTARY_PROFILE)" --wait
	xcrun stapler staple $(DIST)/Qalam-$(VERSION).pkg
endif
	# The same package under a name that never changes, for the website's download button.
	cp $(DIST)/Qalam-$(VERSION).pkg $(DIST)/Qalam-mac.pkg
	sed -i '' -E "s/^  version \".*\"/  version \"$(VERSION)\"/; s/^  sha256 \".*\"/  sha256 \"$$(shasum -a 256 $(DIST)/Qalam-mac.zip | cut -d' ' -f1)\"/" Casks/qalam.rb

release: mac-release windows
	cd $(DIST) && shasum -a 256 * > SHA256SUMS.txt
	@ls -la $(DIST)

# ---------- Website ----------

# The live demo on the website runs the Windows engine (Go) compiled to WebAssembly.
wasm:
	mkdir -p website/demo
	cd windows && GOOS=js GOARCH=wasm $(GO) build -trimpath -ldflags "-s -w" -o ../website/demo/qalam.wasm ./wasm
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" website/demo/wasm_exec.js

# build/site is what GitHub Pages serves: the website, the docs (the wiki) and the key chart.
site: wasm
	rm -rf $(BUILD)/site
	mkdir -p $(BUILD)/site
	cp -R website/. $(BUILD)/site/
	cp -R docs $(BUILD)/site/docs
	cp LAYOUT.md $(BUILD)/site/docs/LAYOUT.md
	mkdir -p $(BUILD)/site/fonts
	cp fonts/KFGQPC-HAFS-Uthmanic-Script.ttf fonts/LICENSE-KFGQPC.txt $(BUILD)/site/fonts/
	@echo "Preview: python3 -m http.server -d $(BUILD)/site 8000"

clean:
	rm -rf .build $(BUILD) $(DIST)
