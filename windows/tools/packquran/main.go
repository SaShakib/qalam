// packquran copies the Qur'an data (../quran) into assets/quran for the Windows app to embed:
// the .tsv files and README as they are, the texts gzipped (about 1.3 MB instead of 6.8 MB).
//
//	go run ./tools/packquran ../quran assets/quran
package main

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: packquran <quran folder> <output folder>")
		os.Exit(2)
	}
	src, dst := os.Args[1], os.Args[2]
	if err := pack(src, dst); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func pack(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		data, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			return err
		}
		switch {
		case strings.HasSuffix(name, ".txt"):
			f, err := os.Create(filepath.Join(dst, name+".gz"))
			if err != nil {
				return err
			}
			z, _ := gzip.NewWriterLevel(f, gzip.BestCompression) // no name or time: the same bytes every run
			if _, err := z.Write(data); err != nil {
				return err
			}
			if err := z.Close(); err != nil {
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		case strings.HasSuffix(name, ".tsv"), name == "README.md":
			if err := os.WriteFile(filepath.Join(dst, name), data, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}
