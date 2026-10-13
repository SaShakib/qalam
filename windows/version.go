package main

import (
	"runtime"
	"strconv"
	"strings"
)

var version = "dev" // set at build time: -X main.version=…

// assetName is the release file for this computer's processor.
func assetName() string {
	if runtime.GOARCH == "arm64" {
		return "Qalam-Windows-arm64.exe"
	}
	return "Qalam-Windows-x64.exe"
}

// isNewer compares dotted versions: isNewer("1.10.0", "1.9.2") is true.
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
