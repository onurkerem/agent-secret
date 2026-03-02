package parser

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestNeedsQuoting(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"simple", false},
		{"with space", true},
		{"with!special", true},
		{"password@", true},
		{"my#password", true},
		{"value$var", true},
		{"a%b", true},
		{"a^b", true},
		{"a&b", true},
		{"a*b", true},
		{"a(b)c", true},
		{"a|b", true},
		{"a<b", true},
		{"a>b", true},
		{"a?b", true},
		{"a;b", true},
		{"a~b", true},
		{`a\b`, true},
		{`"quoted"`, true},
		{"'single'", true},
		{"123", false},
		{"abc123", false},
		{"ABC_123", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := needsQuoting(tt.value)
			if result != tt.expected {
				t.Errorf("needsQuoting(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestEscapeValue(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		{`simple`, `simple`},
		{`with"quote`, `with\"quote`},
		{`with\backslash`, `with\\backslash`},
		{`both"and\`, `both\"and\\`},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := escapeValue(tt.value)
			if result != tt.expected {
				t.Errorf("escapeValue(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		{"simple", "simple"},
		{"with space", `"with space"`},
		{`with"quote`, `"with\"quote"`},
		{"password!", `"password!"`},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := formatValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatValue(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		line        string
		expectedKey string
		expectedVal string
		isComment   bool
	}{
		{"KEY=value", "KEY", "value", false},
		{"KEY = value", "KEY", "value", false},
		{"KEY=", "KEY", "", false},
		{"KEY=\"quoted value\"", "KEY", "quoted value", false},
		{"KEY='single quoted'", "KEY", "single quoted", false},
		{"# This is a comment", "", "", true},
		{"  # Indented comment", "", "", true},
		{"", "", "", false},
		{"   ", "", "", false},
		{"NO_EQUALS_HERE", "NO_EQUALS_HERE", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			key, val, isComment := parseLine(tt.line)
			if key != tt.expectedKey || val != tt.expectedVal || isComment != tt.isComment {
				t.Errorf("parseLine(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.line, key, val, isComment,
					tt.expectedKey, tt.expectedVal, tt.isComment)
			}
		})
	}
}

func TestInjectSecretsNewFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	secrets := map[string]string{
		"API_KEY":     "secret123",
		"DB_PASSWORD": "my p@ssword!",
	}

	injected, err := InjectSecrets(envFile, secrets)
	if err != nil {
		t.Fatalf("InjectSecrets failed: %v", err)
	}

	if len(injected) != 2 {
		t.Errorf("Expected 2 injected secrets, got %d", len(injected))
	}

	// Read and verify file content
	content, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Check both secrets are present (order is non-deterministic due to map)
	if !contains(contentStr, "API_KEY=secret123") {
		t.Error("API_KEY not found in file")
	}
	if !contains(contentStr, `DB_PASSWORD="my p@ssword!"`) {
		t.Error("DB_PASSWORD not found in file with proper quoting")
	}
}

func TestInjectSecretsExistingFile(t *testing.T) {
	// Create temp directory with existing .env
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	// Create initial file
	initialContent := `# Database configuration
DB_HOST=localhost
DB_PORT=5432

# API Keys
OLD_KEY=old_value
`
	if err := os.WriteFile(envFile, []byte(initialContent), 0600); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	secrets := map[string]string{
		"DB_HOST":     "newhost",
		"NEW_SECRET":  "new_value",
		"SPECIAL_VAL": "has spaces!",
	}

	injected, err := InjectSecrets(envFile, secrets)
	if err != nil {
		t.Fatalf("InjectSecrets failed: %v", err)
	}

	// Should have injected 3 secrets
	if len(injected) != 3 {
		t.Errorf("Expected 3 injected secrets, got %d", len(injected))
	}

	// Read and verify file content
	content, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Check that comments and structure are preserved
	contentStr := string(content)
	if !contains(contentStr, "# Database configuration") {
		t.Error("Comment 'Database configuration' was not preserved")
	}
	if !contains(contentStr, "# API Keys") {
		t.Error("Comment 'API Keys' was not preserved")
	}
	if !contains(contentStr, "DB_HOST=newhost") {
		t.Error("DB_HOST was not updated")
	}
	if !contains(contentStr, "DB_PORT=5432") {
		t.Error("DB_PORT was not preserved")
	}
	if !contains(contentStr, "OLD_KEY=old_value") {
		t.Error("OLD_KEY was not preserved")
	}
	if !contains(contentStr, "NEW_SECRET=new_value") {
		t.Error("NEW_SECRET was not appended")
	}
	if !contains(contentStr, `SPECIAL_VAL="has spaces!"`) {
		t.Error("SPECIAL_VAL was not properly formatted with quotes")
	}
}

func TestInjectSecretsPreservesWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	initialContent := `KEY1=value1

KEY2=value2


KEY3=value3
`
	if err := os.WriteFile(envFile, []byte(initialContent), 0600); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	secrets := map[string]string{
		"KEY1": "updated1",
		"KEY4": "new4",
	}

	_, err := InjectSecrets(envFile, secrets)
	if err != nil {
		t.Fatalf("InjectSecrets failed: %v", err)
	}

	content, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Check blank lines are preserved
	if !contains(contentStr, "KEY1=updated1\n\nKEY2=value2") {
		t.Error("Blank line between KEY1 and KEY2 was not preserved")
	}
	if !contains(contentStr, "KEY2=value2\n\n\nKEY3=value3") {
		t.Error("Multiple blank lines between KEY2 and KEY3 were not preserved")
	}
}

func TestParseEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `# Comment
KEY1=value1
KEY2="quoted value"
KEY3='single quoted'

# Another comment
KEY4=simple
`
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	result, err := ParseEnvFile(envFile)
	if err != nil {
		t.Fatalf("ParseEnvFile failed: %v", err)
	}

	expected := map[string]string{
		"KEY1": "value1",
		"KEY2": "quoted value",
		"KEY3": "single quoted",
		"KEY4": "simple",
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d keys, got %d", len(expected), len(result))
	}

	for k, v := range expected {
		if result[k] != v {
			t.Errorf("result[%q] = %q, want %q", k, result[k], v)
		}
	}
}

func TestParseEnvFileNotExists(t *testing.T) {
	result, err := ParseEnvFile("/nonexistent/.env")
	if err != nil {
		t.Errorf("ParseEnvFile should not error for nonexistent file, got: %v", err)
	}
	if result == nil {
		t.Error("Result should be non-nil empty map")
	}
	if len(result) != 0 {
		t.Errorf("Expected empty map, got %d keys", len(result))
	}
}

func TestValidateEnvKey(t *testing.T) {
	tests := []struct {
		key       string
		shouldErr bool
	}{
		{"VALID_KEY", false},
		{"valid_key", false},
		{"ValidKey123", false},
		{"_UNDERSCORE", false},
		{"123INVALID", true},
		{"", true},
		{"INVALID-KEY", true},
		{"INVALID.KEY", true},
		{"INVALID KEY", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := ValidateEnvKey(tt.key)
			if tt.shouldErr && err == nil {
				t.Errorf("ValidateEnvKey(%q) should have returned error", tt.key)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("ValidateEnvKey(%q) should not have returned error, got: %v", tt.key, err)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure sorted output for consistent test results
func sortStrings(s []string) {
	sort.Strings(s)
}
