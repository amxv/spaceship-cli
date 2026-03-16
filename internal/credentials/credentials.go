package credentials

import (
	"errors"
	"os"
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

	key, err := loadKeychain(apiKeyAccount)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", "", ErrNotFound
		}
		return "", "", err
	}

	secret, err = loadKeychain(apiSecretAccount)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", "", ErrNotFound
		}
		return "", "", err
	}

	return key, secret, nil
}

func Save(key, secret string) error {
	if err := saveKeychain(apiKeyAccount, key); err != nil {
		return err
	}
	if err := saveKeychain(apiSecretAccount, secret); err != nil {
		return err
	}
	return nil
}

func Delete() error {
	if err := deleteKeychain(apiKeyAccount); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if err := deleteKeychain(apiSecretAccount); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return nil
}
