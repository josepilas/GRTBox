package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunCommandReturnsOutputForNonZeroExit(t *testing.T) {
	command := "sh"
	args := []string{"-c", "echo stdout-message; echo stderr-message >&2; exit 7"}
	if runtime.GOOS == "windows" {
		command = "powershell.exe"
		args = []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", "Write-Output 'stdout-message'; Write-Error 'stderr-message'; exit 7"}
	}

	result, err := runCommand(command, args, RuntimeExecOptions{TimeoutSeconds: 10})
	if err != nil {
		t.Fatalf("expected non-zero command result without Go error, got %v", err)
	}
	if result.ExitCode != 7 {
		t.Fatalf("expected exit code 7, got %d", result.ExitCode)
	}
	if !strings.Contains(result.Stdout, "stdout-message") {
		t.Fatalf("expected stdout to be preserved, got %q", result.Stdout)
	}
	if !strings.Contains(result.Stderr, "stderr-message") {
		t.Fatalf("expected stderr to be preserved, got %q", result.Stderr)
	}
}

func TestRuntimeCryptoHashFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hash.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	hash, err := app.RuntimeCryptoHashFile(path, "sha256")
	if err != nil {
		t.Fatal(err)
	}
	if hash != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("unexpected sha256: %s", hash)
	}
}

func TestRuntimeFilesystemAndPathHelpers(t *testing.T) {
	app := NewApp()
	dir := t.TempDir()
	file := app.RuntimePathJoin([]string{dir, "nested", "file.txt"})

	if err := app.RuntimeFilesystemMkdirAll(app.RuntimePathDirName(file)); err != nil {
		t.Fatal(err)
	}
	if err := app.RuntimeFilesystemWriteFile(file, "content"); err != nil {
		t.Fatal(err)
	}

	info := app.RuntimeFilesystemStat(file)
	if !info.Exists || info.IsDir || info.Size != 7 {
		t.Fatalf("unexpected file info: %#v", info)
	}
	if app.RuntimePathBaseName(file) != "file.txt" {
		t.Fatalf("unexpected basename: %s", app.RuntimePathBaseName(file))
	}
	if app.RuntimePathExtName(file) != ".txt" {
		t.Fatalf("unexpected extension: %s", app.RuntimePathExtName(file))
	}
}
