package keyring

import (
	"strings"

	"github.com/zalando/go-keyring"
)

// Set stores a secret in the OS keychain
func Set(service, name, value string) error {
	return keyring.Set(service, name, value)
}

// Get retrieves a secret from the OS keychain
func Get(service, name string) (string, error) {
	return keyring.Get(service, name)
}

// Delete removes a secret from the OS keychain
func Delete(service, name string) error {
	return keyring.Delete(service, name)
}

// Exists checks if a secret exists in the OS keychain
func Exists(service, name string) bool {
	_, err := keyring.Get(service, name)
	return err == nil
}

// List returns all secret names stored under a service
// Note: This is a workaround since go-keyring doesn't have a native List function
// We maintain a separate index of stored secrets
func List(service string) ([]string, error) {
	// The go-keyring library doesn't support listing all secrets directly
	// We'll use a workaround by storing a list of secret names as a metadata entry
	indexKey := "__agent_secret_index__"
	indexValue, err := keyring.Get(service, indexKey)
	if err != nil {
		// No index exists yet, return empty list
		return []string{}, nil
	}

	if indexValue == "" {
		return []string{}, nil
	}

	// Split by newline
	var names []string
	for _, name := range splitString(indexValue, "\n") {
		if name != "" && name != indexKey {
			names = append(names, name)
		}
	}
	return names, nil
}

// addToIndex adds a secret name to the index
func addToIndex(service, name string) error {
	indexKey := "__agent_secret_index__"
	existingIndex, _ := keyring.Get(service, indexKey)

	// Check if already in index
	for _, n := range splitString(existingIndex, "\n") {
		if n == name {
			return nil // Already exists
		}
	}

	// Add to index
	newIndex := existingIndex
	if newIndex != "" {
		newIndex += "\n"
	}
	newIndex += name

	return keyring.Set(service, indexKey, newIndex)
}

// removeFromIndex removes a secret name from the index
func removeFromIndex(service, name string) error {
	indexKey := "__agent_secret_index__"
	existingIndex, err := keyring.Get(service, indexKey)
	if err != nil {
		return nil // No index to update
	}

	var newNames []string
	for _, n := range splitString(existingIndex, "\n") {
		if n != "" && n != name {
			newNames = append(newNames, n)
		}
	}

	if len(newNames) == 0 {
		return keyring.Delete(service, indexKey)
	}

	newIndex := ""
	for i, n := range newNames {
		if i > 0 {
			newIndex += "\n"
		}
		newIndex += n
	}

	return keyring.Set(service, indexKey, newIndex)
}

// Set stores a secret in the OS keychain and updates the index
func SetWithIndex(service, name, value string) error {
	if err := keyring.Set(service, name, value); err != nil {
		return err
	}
	return addToIndex(service, name)
}

// Delete removes a secret from the OS keychain and updates the index
func DeleteWithIndex(service, name string) error {
	if err := keyring.Delete(service, name); err != nil {
		return err
	}
	return removeFromIndex(service, name)
}

// Helper function to split strings
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// IsKeyringError checks if the error is related to keyring unavailability
// (e.g., headless Linux environment without GUI)
func IsKeyringError(err error) bool {
	// Check for common keyring error patterns
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Common error messages from go-keyring when keyring is unavailable
	return strings.Contains(errStr, "keyring") ||
		strings.Contains(errStr, "secret service") ||
		strings.Contains(errStr, "dbus")
}
