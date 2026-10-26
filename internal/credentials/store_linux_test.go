//go:build linux

package credentials

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupLinuxCredentialsTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv(envAPIKey, "")
	t.Setenv(envAPISecret, "")
	return filepath.Join(root, "spaceship-cli", "credentials.json")
}

func TestLinuxCredentialsLifecycle(t *testing.T) {
	filePath := setupLinuxCredentialsTest(t)
	if _, _, err := Load(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing Load error: %v", err)
	}
	if err := Save("key-1", "secret-1"); err != nil {
		t.Fatal(err)
	}
	checkMode := func(path string, expected os.FileMode) {
		t.Helper()
		info, err := os.Lstat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != expected {
			t.Errorf("%s permissions = %o, want %o", path, got, expected)
		}
	}
	checkMode(filepath.Dir(filePath), 0o700)
	checkMode(filePath, 0o600)
	if key, secret, err := Load(); err != nil || key != "key-1" || secret != "secret-1" {
		t.Fatalf("Load = (%q, %q, %v)", key, secret, err)
	}
	if err := Save("key-2", "secret-2"); err != nil {
		t.Fatal(err)
	}
	checkMode(filePath, 0o600)
	if key, secret, err := Load(); err != nil || key != "key-2" || secret != "secret-2" {
		t.Fatalf("replaced Load = (%q, %q, %v)", key, secret, err)
	}
	if err := Delete(); err != nil {
		t.Fatal(err)
	}
	if err := Delete(); err != nil { // logout is idempotent
		t.Fatal(err)
	}
	if _, err := os.Lstat(filePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected removed file: %v", err)
	}
}

func TestLinuxEnvironmentCredentialsTakePrecedence(t *testing.T) {
	setupLinuxCredentialsTest(t)
	if err := Save("stored-key", "stored-secret"); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAPIKey, "environment-key")
	t.Setenv(envAPISecret, "environment-secret")
	if key, secret, err := Load(); err != nil || key != "environment-key" || secret != "environment-secret" {
		t.Fatalf("environment Load = (%q, %q, %v)", key, secret, err)
	}
	t.Setenv(envAPISecret, "")
	if key, secret, err := Load(); err != nil || key != "stored-key" || secret != "stored-secret" {
		t.Fatalf("partial environment Load = (%q, %q, %v)", key, secret, err)
	}
}

func TestLinuxCredentialsRejectInsecurePaths(t *testing.T) {
	t.Run("world-readable credentials", func(t *testing.T) {
		filePath := setupLinuxCredentialsTest(t)
		if err := Save("safe", "safe"); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(filePath, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, _, err := Load(); err == nil || !strings.Contains(err.Error(), "chmod 600") {
			t.Fatalf("should reject unsafe credentials permissions: %v", err)
		}
		if err := Save("should", "fail"); err == nil {
			t.Fatal("Save should reject unsafe existing file")
		}
	})
	t.Run("world-accessible directory", func(t *testing.T) {
		filePath := setupLinuxCredentialsTest(t)
		if err := Save("safe", "safe"); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(filepath.Dir(filePath), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, _, err := Load(); err == nil || !strings.Contains(err.Error(), "chmod 700") {
			t.Fatalf("should reject unsafe directory permissions: %v", err)
		}
	})
	t.Run("symlink credentials file", func(t *testing.T) {
		filePath := setupLinuxCredentialsTest(t)
		if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), "target")
		if err := os.WriteFile(target, []byte("untouched"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filePath); err != nil {
			t.Fatal(err)
		}
		if err := Save("bad", "bad"); err == nil {
			t.Fatal("Save followed symlink")
		}
		if _, _, err := Load(); err == nil {
			t.Fatal("Load followed symlink")
		}
		data, err := os.ReadFile(target)
		if err != nil || string(data) != "untouched" {
			t.Fatalf("symlink target changed: %q, %v", data, err)
		}
	})
	t.Run("symlink directory", func(t *testing.T) {
		filePath := setupLinuxCredentialsTest(t)
		if err := os.Symlink(t.TempDir(), filepath.Dir(filePath)); err != nil {
			t.Fatal(err)
		}
		if err := Save("bad", "bad"); err == nil {
			t.Fatal("Save should reject symlink directory")
		}
	})
	t.Run("incomplete credentials", func(t *testing.T) {
		filePath := setupLinuxCredentialsTest(t)
		if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte(`{"apiKey":"only-key"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := Load(); err == nil || !strings.Contains(err.Error(), "incomplete") {
			t.Fatalf("should reject incomplete pair: %v", err)
		}
	})
}
