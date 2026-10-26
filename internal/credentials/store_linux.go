//go:build linux

package credentials

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Linux deliberately uses an owner-only file rather than a desktop Secret
// Service dependency: the CLI must also work in SSH sessions and containers.
// This protects against other unprivileged users, not against the account
// owner or disk access. Disk encryption is recommended for secrets at rest.
type linuxCredentials struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

func linuxCredentialsPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate Linux config directory: %w", err)
	}
	if !filepath.IsAbs(base) {
		return "", fmt.Errorf("Linux config directory must be an absolute path: %q", base)
	}
	return filepath.Join(base, "spaceship-cli", "credentials.json"), nil
}

// Only the application's private leaf directory is restricted. Do not chmod
// the user's shared ~/.config or any other ancestor directory.
func linuxCredentialsDir(create bool) (string, error) {
	filePath, err := linuxCredentialsPath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(filePath)
	if create {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", fmt.Errorf("create credentials directory: %w", err)
		}
	}
	info, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("inspect credentials directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("credentials directory %s is not a real directory", dir)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return "", fmt.Errorf("insecure permissions on %s: run chmod 700 on this directory", dir)
	}
	return dir, nil
}

func inspectLinuxCredentialsFile(filePath string) error {
	info, err := os.Lstat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNotFound
		}
		return fmt.Errorf("inspect credentials file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("credentials path %s must be a regular file (not a symlink)", filePath)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("insecure permissions on %s: run chmod 600 on this file", filePath)
	}
	return nil
}

func loadStoredCredentials() (string, string, error) {
	if _, err := linuxCredentialsDir(false); err != nil {
		return "", "", err
	}
	filePath, err := linuxCredentialsPath()
	if err != nil {
		return "", "", err
	}
	if err := inspectLinuxCredentialsFile(filePath); err != nil {
		return "", "", err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", fmt.Errorf("read credentials file: %w", err)
	}
	var stored linuxCredentials
	if err := json.Unmarshal(data, &stored); err != nil {
		return "", "", fmt.Errorf("parse credentials file: %w", err)
	}
	if stored.APIKey == "" || stored.APISecret == "" {
		return "", "", errors.New("credentials file is incomplete; run spaceship auth login")
	}
	return stored.APIKey, stored.APISecret, nil
}

func saveStoredCredentials(key, secret string) error {
	dir, err := linuxCredentialsDir(true)
	if err != nil {
		return err
	}
	filePath := filepath.Join(dir, "credentials.json")
	if err := inspectLinuxCredentialsFile(filePath); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	data, err := json.Marshal(linuxCredentials{APIKey: key, APISecret: secret})
	if err != nil {
		return fmt.Errorf("encode credentials: %w", err)
	}
	data = append(data, '\n')
	// CreateTemp uses mode 0600. Rename replaces the entire credential pair
	// atomically, so concurrent readers never see a half-saved login.
	tmp, err := os.CreateTemp(dir, ".credentials-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary credentials file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write credentials: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("flush credentials: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close credentials: %w", err)
	}
	if err := os.Rename(tmp.Name(), filePath); err != nil {
		return fmt.Errorf("replace credentials atomically: %w", err)
	}
	return nil
}

func deleteStoredCredentials() error {
	if _, err := linuxCredentialsDir(false); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	filePath, err := linuxCredentialsPath()
	if err != nil {
		return err
	}
	if err := inspectLinuxCredentialsFile(filePath); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove credentials file: %w", err)
	}
	return nil
}
