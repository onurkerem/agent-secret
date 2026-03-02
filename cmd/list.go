package cmd

import (
	"fmt"

	"github.com/onurkerem/agent-secret/internal/keyring"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored secret names",
	Long: `Lists all secret names stored in the OS keychain under the agent-secret service.

Note: This command only lists the names of secrets, not their values.
To retrieve a value, use the 'get' command.

Example:
  agent-secret list`,
	Run: func(cmd *cobra.Command, args []string) {
		names, err := keyring.List(serviceName)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error listing secrets: %v\n", err)
			return
		}

		if len(names) == 0 {
			fmt.Println("No secrets stored.")
			return
		}

		fmt.Printf("Stored secrets (%d):\n", len(names))
		for _, name := range names {
			fmt.Printf("  - %s\n", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
