package cmd

import (
	"fmt"
	"os"

	"github.com/onurkerem/agent-secret/internal/keyring"
	"github.com/onurkerem/agent-secret/internal/prompt"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <SECRET_NAME>",
	Short: "Store a secret securely in the OS vault",
	Long: `Stores a secret securely in your operating system's native keychain.

The secret value is prompted securely (hidden input) and never appears in
command-line arguments or shell history.

Example:
  agent-secret set API_KEY
  agent-secret set DB_PASSWORD`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		secretName := args[0]

		// Prompt for secret value (hidden input)
		secretValue, err := prompt.PromptSecret(fmt.Sprintf("Enter value for %s: ", secretName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading secret: %v\n", err)
			os.Exit(1)
		}

		if secretValue == "" {
			fmt.Fprintln(os.Stderr, "Error: Secret value cannot be empty")
			os.Exit(1)
		}

		// Store in OS keychain
		if err := keyring.SetWithIndex(serviceName, secretName, secretValue); err != nil {
			fmt.Fprintf(os.Stderr, "Error storing secret: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Secret '%s' stored successfully\n", secretName)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
