package credentials

import (
	"errors"
	"os"
	"runtime"
	"strings"
)

const (
	envAPIKey    = "SPACESHIP_API_KEY"
	envAPISecret = "SPACESHIP_API_SECRET"
)

var ErrNotFound = errors.New("credentials not found")

func Load() (string, string, error) {
	key := strings.TrimSpace(os.Getenv(envAPIKey))
	secret := strings.TrimSpace(os.Getenv(envAPISecret))
	if key != "" && secret != "" {
		return key, secret, nil
	}
	return loadStoredCredentials()
}

func Save(key, secret string) error {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(secret) == "" {
		return errors.New("API key and secret must both be non-empty")
	}
	return saveStoredCredentials(key, secret)
}

func Delete() error {
	return deleteStoredCredentials()
}

// StoredDescription is used for status/help text, not to expose secret values.
func StoredDescription() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS Keychain (service spaceship-cli)"
	case "linux":
		return "Linux credential file (XDG config directory, owner-only)"
	default:
		return "local credential store"
	}
}
