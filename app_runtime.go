//go:build windows || linux

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"wails-go-desktop-code-interactive/internal/runtimebundles"
)

func (a *App) prepareRuntimeBundles() error {
	dir, err := os.MkdirTemp("", "wails-runtime-*")
	if err != nil {
		a.runtimeRoot = ""
		return err
	}

	if err := runtimebundles.Extract(dir); err != nil {
		_ = os.RemoveAll(dir)
		a.runtimeRoot = ""
		if errors.Is(err, runtimebundles.ErrNoBundledRuntimes) {
			return nil
		}
		return err
	}

	a.runtimeRoot = dir
	return nil
}

func (a *App) shutdown(ctx context.Context) {
	a.cleanupRuntimeBundles()
}

func (a *App) cleanupRuntimeBundles() {
	if a.runtimeRoot == "" {
		return
	}
	_ = os.RemoveAll(a.runtimeRoot)
	a.runtimeRoot = ""
}

func (a *App) ListBundledRuntimes() []string {
	if a.runtimeRoot == "" {
		return []string{}
	}

	entries, err := os.ReadDir(a.runtimeRoot)
	if err != nil {
		return []string{}
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

func defaultExecutableName(language string) string {
	name := language
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(strings.ToLower(name), ".exe") {
			name += ".exe"
		}
	}
	return name
}

func (a *App) findBundledExecutable(language string) (string, string) {
	if a.runtimeRoot == "" {
		return "", ""
	}

	execName := defaultExecutableName(language)
	entries, err := os.ReadDir(a.runtimeRoot)
	if err != nil {
		return "", ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(a.runtimeRoot, entry.Name(), execName)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, entry.Name()
		}
	}

	return "", ""
}

func (a *App) resolveBundledExecutable(language, folder string) (string, string, error) {
	if a.runtimeRoot == "" {
		return "", "", fmt.Errorf("no bundled runtimes available")
	}

	execName := defaultExecutableName(language)

	if folder != "" {
		base := filepath.Join(a.runtimeRoot, folder)
		candidate := filepath.Join(base, execName)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, base, nil
		}

		candidate = base
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, filepath.Dir(candidate), nil
		}

		return "", "", fmt.Errorf("bundled runtime %s missing executable for %s", folder, language)
	}

	if candidate, runtimeName := a.findBundledExecutable(language); candidate != "" {
		return candidate, filepath.Join(a.runtimeRoot, runtimeName), nil
	}

	return "", "", fmt.Errorf("no bundled executable found for %s", language)
}

func (a *App) resolveCustomExecutable(language, customPath string) (string, string, error) {
	if customPath == "" {
		return "", "", fmt.Errorf("custom executable path is empty")
	}

	resolved := customPath
	if !filepath.IsAbs(resolved) {
		if a.runtimeRoot != "" {
			candidate := filepath.Join(a.runtimeRoot, resolved)
			if _, err := os.Stat(candidate); err == nil {
				resolved = candidate
			} else if abs, err := filepath.Abs(resolved); err == nil {
				resolved = abs
			} else {
				return "", "", err
			}
		} else if abs, err := filepath.Abs(resolved); err == nil {
			resolved = abs
		} else {
			return "", "", err
		}
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", "", err
	}

	if info.IsDir() {
		candidate := filepath.Join(resolved, defaultExecutableName(language))
		if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
			return candidate, resolved, nil
		}
		return "", "", fmt.Errorf("executable %s not found in %s", defaultExecutableName(language), resolved)
	}

	return resolved, filepath.Dir(resolved), nil
}

func (a *App) resolveExecutionTarget(data Data) (string, string, error) {
	mode := strings.ToLower(strings.TrimSpace(data.ExecMode))
	switch mode {
	case "bundled":
		cmd, dir, err := a.resolveBundledExecutable(data.Language, strings.TrimSpace(data.BundledRuntime))
		if err != nil {
			return "", "", err
		}
		return cmd, dir, nil
	case "custom":
		cmd, dir, err := a.resolveCustomExecutable(data.Language, strings.TrimSpace(data.CustomExecutable))
		if err != nil {
			return "", "", err
		}

		if data.CustomWorkingDir != "" {
			workingDir := data.CustomWorkingDir
			if !filepath.IsAbs(workingDir) {
				workingDir = filepath.Join(dir, workingDir)
			}
			return cmd, workingDir, nil
		}
		return cmd, dir, nil
	default:
		if data.PreferBundled {
			if cmd, dir, err := a.resolveBundledExecutable(data.Language, strings.TrimSpace(data.BundledRuntime)); err == nil {
				return cmd, dir, nil
			}
		}
		return data.Language, "", nil
	}
}

func (a *App) hasBundledExecutable(language string) bool {
	_, _, err := a.resolveBundledExecutable(language, "")
	return err == nil
}
