//go:build !darwin && !linux

package credentials

import "errors"

const (
	serviceName      = "spaceship-cli"
	apiKeyAccount    = "api-key"
	apiSecretAccount = "api-secret"
)

func saveKeychain(account, value string) error {
	return errors.New("credential storage is only implemented for macOS and Linux; set SPACESHIP_API_KEY and SPACESHIP_API_SECRET")
}

func loadKeychain(account string) (string, error) {
	return "", ErrNotFound
}

func deleteKeychain(account string) error {
	return ErrNotFound
}
