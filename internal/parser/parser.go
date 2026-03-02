package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// needsQuoting checks if a value needs to be wrapped in double quotes
func needsQuoting(value string) bool {
	// Check for spaces
	if strings.Contains(value, " ") {
		return true
	}
	// Check for special characters that would break .env parsing
	specialChars := "!@#$%^&*()|<>?;~\\`\""
	for _, char := range specialChars {
		if strings.Contains(value, string(char)) {
			return true
		}
	}
	// Check if it starts or ends with quotes (needs escaping)
	if strings.HasPrefix(value, `"`) || strings.HasSuffix(value, `"`) {
		return true
	}
	if strings.HasPrefix(value, "'") || strings.HasSuffix(value, "'") {
		return true
	}
	return false
}

// escapeValue escapes double quotes within a value
func escapeValue(value string) string {
	// Escape backslashes first, then double quotes
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

// formatValue formats a secret value for .env file, wrapping in quotes if necessary
func formatValue(value string) string {
	if needsQuoting(value) {
		return `"` + escapeValue(value) + `"`
	}
	return value
}

// parseLine parses a .env line into key and value
// Returns empty key for comments and empty lines
func parseLine(line string) (key string, value string, isComment bool) {
	trimmed := strings.TrimSpace(line)

	// Empty line
	if trimmed == "" {
		return "", "", false
	}

	// Comment line
	if strings.HasPrefix(trimmed, "#") {
		return "", "", true
	}

	// Find the first = sign
	eqIndex := strings.Index(trimmed, "=")
	if eqIndex == -1 {
		// No = sign, treat as key with empty value
		return trimmed, "", false
	}

	key = strings.TrimSpace(trimmed[:eqIndex])
	value = strings.TrimSpace(trimmed[eqIndex+1:])

	// Remove surrounding quotes if present
	if len(value) >= 2 {
		if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
			(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
			value = value[1 : len(value)-1]
		}
	}

	return key, value, false
}

// InjectSecrets injects secrets into a .env file
// Returns the list of secrets that were injected/updated
func InjectSecrets(filePath string, secrets map[string]string) ([]string, error) {
	// Read existing file if it exists
	var lines []string
	existingFile, err := os.Open(filePath)
	if err == nil {
		scanner := bufio.NewScanner(existingFile)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		existingFile.Close()
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	// Track which secrets exist in the file
	secretsInFile := make(map[string]bool)
	updatedSecrets := make(map[string]bool)

	// First pass: update existing secrets
	for i, line := range lines {
		key, _, isComment := parseLine(line)
		if isComment || key == "" {
			continue
		}

		if newValue, exists := secrets[key]; exists {
			// Update this line with the new value
			lines[i] = key + "=" + formatValue(newValue)
			secretsInFile[key] = true
			updatedSecrets[key] = true
		}
	}

	// Second pass: append new secrets that don't exist
	var newSecrets []string
	for key, value := range secrets {
		if !secretsInFile[key] {
			newSecrets = append(newSecrets, key+"="+formatValue(value))
			updatedSecrets[key] = true
		}
	}

	// Append new secrets to the end
	if len(newSecrets) > 0 {
		// Add a blank line if file is not empty and doesn't end with blank line
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, newSecrets...)
	}

	// Write the file
	output := strings.Join(lines, "\n")
	if len(lines) > 0 {
		output += "\n"
	}

	if err := os.WriteFile(filePath, []byte(output), 0600); err != nil {
		return nil, fmt.Errorf("error writing file: %w", err)
	}

	// Return list of injected secrets
	var result []string
	for key := range updatedSecrets {
		result = append(result, key)
	}
	return result, nil
}

// ParseEnvFile reads a .env file and returns a map of key-value pairs
// This is useful for reading existing values
func ParseEnvFile(filePath string) (map[string]string, error) {
	result := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil // Return empty map if file doesn't exist
		}
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		key, value, isComment := parseLine(line)
		if isComment || key == "" {
			continue
		}
		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return result, nil
}

// ValidateEnvKey validates that a key is a valid .env variable name
func ValidateEnvKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Must start with a letter or underscore
	first := key[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return fmt.Errorf("key must start with a letter or underscore")
	}

	// Can only contain letters, digits, and underscores
	validKey := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !validKey.MatchString(key) {
		return fmt.Errorf("key can only contain letters, digits, and underscores")
	}

	return nil
}
