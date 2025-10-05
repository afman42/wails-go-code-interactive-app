package runtimebundles

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed runtimes/**
var runtimeFS embed.FS

var ErrNoBundledRuntimes = errors.New("no bundled runtimes available")

func Extract(target string) error {
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}

	hasFiles := false

	err := fs.WalkDir(runtimeFS, "runtimes", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "runtimes" {
			return nil
		}

		rel, err := filepath.Rel("runtimes", path)
		if err != nil {
			return err
		}

		destination := filepath.Join(target, rel)

		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}

		hasFiles = true

		data, err := runtimeFS.ReadFile(path)
		if err != nil {
			return err
		}

		mode := fs.FileMode(0o755)
		if info, err := entry.Info(); err == nil {
			mode = info.Mode()
			if mode&0o111 == 0 {
				mode |= 0o111
			}
		}

		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}

		return os.WriteFile(destination, data, mode)
	})

	if err != nil {
		return err
	}

	if !hasFiles {
		return ErrNoBundledRuntimes
	}

	return nil
}

func ListDirectories() ([]string, error) {
	entries, err := runtimeFS.ReadDir("runtimes")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}
