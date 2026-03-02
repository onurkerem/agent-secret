package cmd

import (
	"fmt"
	"os"

	"github.com/onurkerem/agent-secret/internal/parser"
	"github.com/spf13/cobra"
)

var (
	checkEnvFile string
	checkQuiet   bool
)

var checkCmd = &cobra.Command{
	Use:   "check <KEY_NAME> [--file | -f <path>]",
	Short: "Check if a key has a non-empty value in an env file",
	Long: `Checks if a key exists and has a non-empty value in a .env file.

This command is designed for use by automated agents that need to verify
secrets are properly configured without being able to see the actual values.

Exit codes:
  0 - Key exists and has a non-empty value
  1 - Key is missing or has an empty value
  2 - Error occurred (file not found, etc.)

Examples:
  agent-secret check DATABASE_URL
  agent-secret check API_KEY -f ./config/.env
  agent-secret check SECRET_TOKEN --quiet`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyName := args[0]

		// Parse the env file
		values, err := parser.ParseEnvFile(checkEnvFile)
		if err != nil {
			if !checkQuiet {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			}
			os.Exit(2)
		}

		// Check if key exists and has non-empty value
		value, exists := values[keyName]
		if !exists {
			if !checkQuiet {
				fmt.Fprintf(os.Stderr, "Key '%s' not found in %s\n", keyName, checkEnvFile)
			}
			os.Exit(1)
		}

		if value == "" {
			if !checkQuiet {
				fmt.Fprintf(os.Stderr, "Key '%s' exists but has an empty value\n", keyName)
			}
			os.Exit(1)
		}

		// Key exists and has a value
		if !checkQuiet {
			fmt.Printf("✓ Key '%s' is set (length: %d characters)\n", keyName, len(value))
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVarP(&checkEnvFile, "file", "f", "./.env", "Target .env file path")
	checkCmd.Flags().BoolVarP(&checkQuiet, "quiet", "q", false, "Suppress output (use exit code only)")
}
