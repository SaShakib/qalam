//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Updates come from GitHub Releases. A running .exe can be renamed but not overwritten, so:
// download Qalam.new.exe → rename Qalam.exe to Qalam.old.exe → move the new one into place →
// start it with --updated → quit. The new copy deletes Qalam.old.exe when it starts.

const releasesAPI = "https://api.github.com/repos/SaShakib/qalam/releases/latest"

var updateMu sync.Mutex

type ghRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func (r ghRelease) version() string { return strings.TrimPrefix(r.TagName, "v") }

func assetName() string {
	if runtime.GOARCH == "arm64" {
		return "Qalam-Windows-arm64.exe"
	}
	return "Qalam-Windows-x64.exe"
}

func isNewer(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		x, y := 0, 0
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			return x > y
		}
	}
	return false
}

func fetchLatest() (ghRelease, error) {
	var r ghRelease
	req, _ := http.NewRequest("GET", releasesAPI, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Qalam/"+version)
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return r, fmt.Errorf("GitHub answered %s", resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r, err
}

// startUpdateLoop checks shortly after launch and then once a day.
func startUpdateLoop() {
	go func() {
		time.Sleep(15 * time.Second)
		for {
			if !settings.NoAutoUpdate {
				checkForUpdate(false)
			}
			time.Sleep(24 * time.Hour)
		}
	}()
}

// checkForUpdate runs on a background goroutine. userAsked shows "up to date" / errors in a dialog.
func checkForUpdate(userAsked bool) {
	if !updateMu.TryLock() {
		return
	}
	defer updateMu.Unlock()

	rel, err := fetchLatest()
	if err != nil {
		if userAsked {
			messageBox("Could not check for updates:\n"+err.Error(), "Qalam", mbIconInfo)
		}
		return
	}
	if version == "dev" || !isNewer(rel.version(), version) {
		if userAsked {
			messageBox("You have the latest version ("+version+").", "Qalam", mbIconInfo)
		}
		return
	}
	if userAsked {
		if messageBox(fmt.Sprintf("Qalam %s is available (you have %s).\n\nInstall it now? Qalam restarts in a few seconds.",
			rel.version(), version), "Qalam update", mbYesNo|mbIconQuestion) != idYes {
			return
		}
	} else {
		balloon("Updating Qalam", "Installing version "+rel.version()+"…")
	}
	if err := applyUpdate(rel); err != nil {
		messageBox("The update could not be installed:\n"+err.Error(), "Qalam", mbIconInfo)
	}
}

func applyUpdate(rel ghRelease) error {
	url := ""
	for _, a := range rel.Assets {
		if a.Name == assetName() {
			url = a.URL
		}
	}
	if url == "" {
		return fmt.Errorf("the release has no %s", assetName())
	}
	exe := exePath()
	dir := filepath.Dir(exe)
	newPath := filepath.Join(dir, "Qalam.new.exe")
	oldPath := filepath.Join(dir, "Qalam.old.exe")

	if err := download(url, newPath); err != nil {
		os.Remove(newPath)
		return err
	}
	os.Remove(oldPath)
	if err := os.Rename(exe, oldPath); err != nil {
		os.Remove(newPath)
		return err
	}
	if err := os.Rename(newPath, exe); err != nil {
		os.Rename(oldPath, exe) // put the old one back
		return err
	}
	if err := exec.Command(exe, "--updated").Start(); err != nil {
		return err
	}
	pPostMessageW.Call(trayHwnd, wmClose, 0, 0) // quit; the new copy takes over
	return nil
}

func download(url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		return err
	}
	if n < 500_000 {
		return fmt.Errorf("the download looks incomplete (%d bytes)", n)
	}
	return nil
}

// cleanupAfterUpdate removes the previous exe left by applyUpdate.
func cleanupAfterUpdate() {
	old := filepath.Join(filepath.Dir(exePath()), "Qalam.old.exe")
	for i := 0; i < 10; i++ {
		if err := os.Remove(old); err == nil || os.IsNotExist(err) {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}
