//go:build !windows

package main

import "fmt"

// The keyboard itself only runs on Windows; on other systems this just explains that.
func main() {
	fmt.Println("Qalam for Windows: build with GOOS=windows (see the Makefile's `windows` target).")
}
