package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/onurkerem/agent-secret/internal/keyring"
	"github.com/onurkerem/agent-secret/internal/parser"
	"github.com/spf13/cobra"
)

var (
	envFile string
)

var injectCmd = &cobra.Command{
	Use:   "inject <SECRET_SPEC_1> [SECRET_SPEC_2] ...",
	Short: "Inject secrets from OS vault into a .env file",
	Long: `Retrieves one or more secrets from the OS vault and writes them into a .env file.

Each argument can be either:
  - A secret name (uses the same name as the key in .env)
  - A mapping in the format SECRET_NAME:KEY_NAME (uses KEY_NAME in .env)

The command validates that all requested secrets exist before modifying any file.
Existing values are updated in-place, and new secrets are appended to the end.
Comments and whitespace in the .env file are preserved.

Examples:
  # Simple: secret name = key name
  agent-secret inject API_KEY
  agent-secret inject API_KEY DB_PASSWORD

  # Mapped: secret name -> different key name
  agent-secret inject PROJECTX_JWT_SECRET:JWT_SECRET
  agent-secret inject PROJECTX_DB_PASS:DB_PASSWORD API_KEY:API_KEY

  # Mixed
  agent-secret inject PROJECTX_JWT_SECRET:JWT_SECRET API_KEY DB_PASS:DATABASE_URL

  # Custom file path
  agent-secret inject API_KEY -f ./config/.env`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse secret specs (SECRET_NAME or SECRET_NAME:KEY_NAME)
		type secretMapping struct {
			secretName string
			keyName    string
		}
		var mappings []secretMapping

		for _, arg := range args {
			if strings.Contains(arg, ":") {
				parts := strings.SplitN(arg, ":", 2)
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					fmt.Fprintf(os.Stderr, "Error: Invalid mapping format '%s'. Use SECRET_NAME:KEY_NAME\n", arg)
					os.Exit(1)
				}
				mappings = append(mappings, secretMapping{
					secretName: parts[0],
					keyName:    parts[1],
				})
			} else {
				// No colon: secret name = key name
				mappings = append(mappings, secretMapping{
					secretName: arg,
					keyName:    arg,
				})
			}
		}

		// First, validate all secrets exist
		secrets := make(map[string]string)
		var missingSecrets []string

		for _, m := range mappings {
			value, err := keyring.Get(serviceName, m.secretName)
			if err != nil {
				missingSecrets = append(missingSecrets, m.secretName)
				continue
			}
			secrets[m.keyName] = value
		}

		if len(missingSecrets) > 0 {
			fmt.Fprintf(os.Stderr, "Error: The following secrets were not found in the keychain:\n")
			for _, name := range missingSecrets {
				fmt.Fprintf(os.Stderr, "  - %s\n", name)
			}
			fmt.Fprintf(os.Stderr, "\nRun 'agent-secret set <SECRET_NAME>' to store them first.\n")
			os.Exit(1)
		}

		// Inject secrets into .env file
		injected, err := parser.InjectSecrets(envFile, secrets)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error injecting secrets: %v\n", err)
			os.Exit(1)
		}

		// Print summary
		fmt.Printf("✓ Successfully injected %d secret(s) into %s\n", len(injected), envFile)
		for _, m := range mappings {
			if m.secretName == m.keyName {
				fmt.Printf("  - %s\n", m.keyName)
			} else {
				fmt.Printf("  - %s (from %s)\n", m.keyName, m.secretName)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(injectCmd)
	injectCmd.Flags().StringVarP(&envFile, "file", "f", "./.env", "Target .env file path")
}
