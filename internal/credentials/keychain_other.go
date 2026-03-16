//go:build !darwin

package credentials

import "errors"

const (
	serviceName      = "spaceship-cli"
	apiKeyAccount    = "api-key"
	apiSecretAccount = "api-secret"
)

func saveKeychain(account, value string) error {
	return errors.New("keychain storage is only implemented for macOS; set SPACESHIP_API_KEY and SPACESHIP_API_SECRET")
}

func loadKeychain(account string) (string, error) {
	return "", ErrNotFound
}

func deleteKeychain(account string) error {
	return ErrNotFound
}
