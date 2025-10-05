package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStringWithCharset(t *testing.T) {
	const length = 10
	const charset = "abcdefgABCDEFG123456"
	for i := 0; i < 5; i++ {
		result := StringWithCharset(length)
		if len(result) != length {
			t.Fatalf("expected length %d, got %d", length, len(result))
		}
		for _, r := range result {
			if !containsRune(charset, r) {
				t.Fatalf("unexpected rune %q in result %q", r, result)
			}
		}
	}
}

func containsRune(set string, r rune) bool {
	for _, candidate := range set {
		if candidate == r {
			return true
		}
	}
	return false
}

func TestMoveFile(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "sample.txt")
	destFile := filepath.Join(destDir, "sample.txt")
	content := []byte("hello world")

	if err := os.WriteFile(srcFile, content, 0o600); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := MoveFile(srcFile, destFile); err != nil {
		t.Fatalf("MoveFile returned error: %v", err)
	}

	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		if err != nil {
			t.Fatalf("unexpected error stating src file: %v", err)
		}
		t.Fatalf("expected source file to be removed")
	}

	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("destination content mismatch: got %q want %q", data, content)
	}
}

func TestPathFileTemp(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	filename := "example.test"
	expected := filepath.Join(wd, "tmp", filename)
	got := PathFileTemp(filename)

	if got != expected {
		t.Fatalf("unexpected path: got %s want %s", got, expected)
	}
}

func TestCheckIsNotData(t *testing.T) {
	slice := []string{"php", "node", "go"}
	if !CheckIsNotData(slice, "php") {
		t.Fatalf("expected value to be found")
	}
	if CheckIsNotData(slice, "python") {
		t.Fatalf("did not expect value to be found")
	}
}
