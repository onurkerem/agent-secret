package cmd

import (
	"fmt"
	"os"

	"github.com/onurkerem/agent-secret/internal/keyring"
	"github.com/onurkerem/agent-secret/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	deleteForce bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <SECRET_NAME>",
	Short: "Remove a secret from the OS vault",
	Long: `Removes a secret from the OS keychain.

By default, asks for confirmation before deleting. Use --force to skip confirmation.

Example:
  agent-secret delete API_KEY
  agent-secret delete OLD_SECRET --force`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		secretName := args[0]

		// Check if secret exists first
		if !keyring.Exists(serviceName, secretName) {
			fmt.Fprintf(os.Stderr, "Error: Secret '%s' not found.\n", secretName)
			os.Exit(1)
		}

		// Confirm deletion unless --force is set
		if !deleteForce {
			confirmed, err := prompt.PromptConfirm(fmt.Sprintf("Delete secret '%s'?", secretName))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading confirmation: %v\n", err)
				os.Exit(1)
			}
			if !confirmed {
				fmt.Println("Cancelled.")
				return
			}
		}

		if err := keyring.DeleteWithIndex(serviceName, secretName); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting secret: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Secret '%s' deleted\n", secretName)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
}
