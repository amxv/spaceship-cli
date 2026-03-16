//go:build darwin

package credentials

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const (
	serviceName      = "spaceship-cli"
	apiKeyAccount    = "api-key"
	apiSecretAccount = "api-secret"
)

func saveKeychain(account, value string) error {
	cmd := exec.Command("security", "add-generic-password", "-a", account, "-s", serviceName, "-w", value, "-U")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to save %s in keychain: %s", account, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func loadKeychain(account string) (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-a", account, "-s", serviceName, "-w")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if strings.Contains(errText, "could not be found") {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to read %s from keychain: %s", account, errText)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func deleteKeychain(account string) error {
	cmd := exec.Command("security", "delete-generic-password", "-a", account, "-s", serviceName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if strings.Contains(errText, "could not be found") {
			return ErrNotFound
		}
		return fmt.Errorf("failed to delete %s from keychain: %s", account, errText)
	}
	return nil
}
