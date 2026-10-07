# Qalam: phonetic Arabic keyboard for macOS and Windows
#   make test        run the transliteration tests (Swift and Go engines)
#   make install     build and install on this Mac (for development)
#   make release     build everything for a GitHub release into dist/
#   make uninstall   remove the development install

VERSION  := 1.1.2
# Builds run at low priority on 2 cores so the Mac stays responsive.
SWIFT    := nice -n 15 swift build -c release -j 2
GO       := nice -n 15 go
RELEASE  := .build/release
BUILD    := build
DIST     := dist
IME_APP  := $(BUILD)/QalamInput.app
APP      := $(BUILD)/Qalam.app
IME_DEST := $(HOME)/Library/Input Methods
APP_DEST := $(HOME)/Applications

.PHONY: all build test icons bundle install uninstall clean guide windows mac-release release

all: bundle

build:
	$(SWIFT)

test:
	$(SWIFT) --product qalam
	$(RELEASE)/qalam --test tests/cases.tsv
	cd windows && $(GO) test ./engine

icons: $(BUILD)/icons/AppIcon.icns

$(BUILD)/icons/AppIcon.icns: tools/make_icons.swift
	nice -n 15 swift tools/make_icons.swift $(BUILD)/icons
	iconutil -c icns $(BUILD)/icons/AppIcon.iconset -o $(BUILD)/icons/AppIcon.icns

bundle: build icons
	rm -rf "$(IME_APP)" "$(APP)"
	mkdir -p "$(IME_APP)/Contents/MacOS" "$(IME_APP)/Contents/Resources/en.lproj"
	cp $(RELEASE)/QalamInput "$(IME_APP)/Contents/MacOS/QalamInput"
	cp mac/bundle/QalamInput-Info.plist "$(IME_APP)/Contents/Info.plist"
	cp mac/bundle/InfoPlist.strings "$(IME_APP)/Contents/Resources/en.lproj/InfoPlist.strings"
	cp $(BUILD)/icons/icon.tiff "$(IME_APP)/Contents/Resources/icon.tiff"
	mkdir -p "$(APP)/Contents/MacOS" "$(APP)/Contents/Resources/Fonts"
	cp $(RELEASE)/QalamApp "$(APP)/Contents/MacOS/Qalam"
	cp mac/bundle/Qalam-Info.plist "$(APP)/Contents/Info.plist"
	cp $(BUILD)/icons/AppIcon.icns "$(APP)/Contents/Resources/AppIcon.icns"
	cp fonts/*.ttf fonts/OFL-*.txt "$(APP)/Contents/Resources/Fonts/"
	for p in "$(IME_APP)" "$(APP)"; do \
	  /usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $(VERSION)" "$$p/Contents/Info.plist"; \
	  /usr/libexec/PlistBuddy -c "Set :CFBundleVersion $(VERSION)" "$$p/Contents/Info.plist"; \
	done
	codesign --force --sign - "$(IME_APP)"
	codesign --force --sign - "$(APP)"

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
	cp $(BUILD)/icons/qalam-on.ico $(BUILD)/icons/qalam-off.ico fonts/NotoNaskhArabic.ttf windows/assets/

windows: guide
	mkdir -p $(DIST)
	cd windows && for arch in amd64 arm64; do \
	  $(GO) run github.com/akavel/rsrc@v0.10.2 -arch $$arch -ico assets/qalam-on.ico -manifest assets/qalam.manifest -o rsrc_windows_$$arch.syso; \
	done
	cd windows && GOOS=windows GOARCH=amd64 $(GO) build -p 2 -trimpath -ldflags "-H windowsgui -s -w -X main.version=$(VERSION)" -o ../$(DIST)/Qalam-Windows-x64.exe .
	cd windows && GOOS=windows GOARCH=arm64 $(GO) build -p 2 -trimpath -ldflags "-H windowsgui -s -w -X main.version=$(VERSION)" -o ../$(DIST)/Qalam-Windows-arm64.exe .

# ---------- Mac release: zip (Homebrew + install.sh) and .pkg (double-click) ----------

mac-release: test bundle
	mkdir -p $(DIST)
	rm -rf $(BUILD)/zip $(BUILD)/pkgroot $(DIST)/Qalam-mac.zip $(DIST)/Qalam-*.pkg
	mkdir -p $(BUILD)/zip
	ditto --norsrc --noextattr "$(APP)" "$(BUILD)/zip/Qalam.app"
	ditto --norsrc --noextattr "$(IME_APP)" "$(BUILD)/zip/QalamInput.app"
	cd $(BUILD)/zip && ditto -c -k --norsrc --noextattr . ../../$(DIST)/Qalam-mac.zip
	mkdir -p "$(BUILD)/pkgroot/Applications" "$(BUILD)/pkgroot/Library/Input Methods"
	ditto --norsrc --noextattr "$(APP)" "$(BUILD)/pkgroot/Applications/Qalam.app"
	ditto --norsrc --noextattr "$(IME_APP)" "$(BUILD)/pkgroot/Library/Input Methods/QalamInput.app"
	pkgbuild --analyze --root $(BUILD)/pkgroot $(BUILD)/components.plist
	for i in 0 1; do /usr/libexec/PlistBuddy -c "Set :$$i:BundleIsRelocatable false" $(BUILD)/components.plist; done
	xattr -cr $(BUILD)/pkgroot
	pkgbuild --root $(BUILD)/pkgroot --component-plist $(BUILD)/components.plist --scripts mac/pkg/scripts \
	  --identifier com.qalam.pkg --version $(VERSION) --install-location / $(DIST)/Qalam-$(VERSION).pkg
	sed -i '' -E "s/^  version \".*\"/  version \"$(VERSION)\"/; s/^  sha256 \".*\"/  sha256 \"$$(shasum -a 256 $(DIST)/Qalam-mac.zip | cut -d' ' -f1)\"/" Casks/qalam.rb

release: mac-release windows
	cd $(DIST) && shasum -a 256 * > SHA256SUMS.txt
	@ls -la $(DIST)

clean:
	rm -rf .build $(BUILD) $(DIST)
