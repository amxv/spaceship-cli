//go:build linux

package credentials

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	serviceName      = "spaceship-cli"
	apiKeyAccount    = "api-key"
	apiSecretAccount = "api-secret"

	linuxCredentialFileName = "credentials.json"
)

type linuxStoredCredentials struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

func linuxCredentialPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to locate user config directory: %w", err)
	}
	return filepath.Join(configDir, serviceName, linuxCredentialFileName), nil
}

func loadLinuxCredentials() (linuxStoredCredentials, error) {
	path, err := linuxCredentialPath()
	if err != nil {
		return linuxStoredCredentials{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return linuxStoredCredentials{}, ErrNotFound
		}
		return linuxStoredCredentials{}, fmt.Errorf("failed to read credentials file: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return linuxStoredCredentials{}, fmt.Errorf("failed to inspect credentials file: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return linuxStoredCredentials{}, fmt.Errorf("credentials file %s must be readable only by the owner", path)
	}

	var stored linuxStoredCredentials
	if err := json.Unmarshal(data, &stored); err != nil {
		return linuxStoredCredentials{}, fmt.Errorf("failed to parse credentials file: %w", err)
	}
	return stored, nil
}

func saveKeychain(account, value string) error {
	stored, err := loadLinuxCredentials()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	switch account {
	case apiKeyAccount:
		stored.APIKey = value
	case apiSecretAccount:
		stored.APISecret = value
	default:
		return fmt.Errorf("unsupported credential account %q", account)
	}

	return writeLinuxCredentials(stored)
}

func loadKeychain(account string) (string, error) {
	stored, err := loadLinuxCredentials()
	if err != nil {
		return "", err
	}

	var value string
	switch account {
	case apiKeyAccount:
		value = stored.APIKey
	case apiSecretAccount:
		value = stored.APISecret
	default:
		return "", fmt.Errorf("unsupported credential account %q", account)
	}
	if value == "" {
		return "", ErrNotFound
	}
	return value, nil
}

func deleteKeychain(account string) error {
	path, err := linuxCredentialPath()
	if err != nil {
		return err
	}

	stored, err := loadLinuxCredentials()
	if err != nil {
		return err
	}

	switch account {
	case apiKeyAccount:
		if stored.APIKey == "" {
			return ErrNotFound
		}
		stored.APIKey = ""
	case apiSecretAccount:
		if stored.APISecret == "" {
			return ErrNotFound
		}
		stored.APISecret = ""
	default:
		return fmt.Errorf("unsupported credential account %q", account)
	}

	if stored.APIKey == "" && stored.APISecret == "" {
		if err := os.Remove(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return ErrNotFound
			}
			return fmt.Errorf("failed to remove credentials file: %w", err)
		}
		return nil
	}

	return writeLinuxCredentials(stored)
}

func writeLinuxCredentials(stored linuxStoredCredentials) error {
	path, err := linuxCredentialPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create credential directory: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("failed to secure credential directory: %w", err)
	}

	data, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("failed to encode credentials: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, ".credentials-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary credentials file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to secure temporary credentials file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to write credentials: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to flush credentials: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temporary credentials file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to install credentials file: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("failed to secure credentials file: %w", err)
	}
	return nil
}
