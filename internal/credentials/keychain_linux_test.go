//go:build linux

package credentials

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxCredentialLifecycle(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv(envAPIKey, "")
	t.Setenv(envAPISecret, "")

	if _, _, err := Load(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Load before Save error = %v, want ErrNotFound", err)
	}

	if err := Save("test-key", "test-secret"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path := filepath.Join(configDir, serviceName, linuxCredentialFileName)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat credentials file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("credentials file permissions = %o, want 600", got)
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat credentials directory: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("credentials directory permissions = %o, want 700", got)
	}

	key, secret, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if key != "test-key" || secret != "test-secret" {
		t.Fatalf("Load() = (%q, %q), want (%q, %q)", key, secret, "test-key", "test-secret")
	}

	if err := Delete(); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("credentials file after Delete error = %v, want not exists", err)
	}
}
