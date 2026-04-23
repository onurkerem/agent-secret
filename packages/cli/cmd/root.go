package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	serviceName = "agent-secret"
	version     = "1.2.0"
)

var rootCmd = &cobra.Command{
	Use:     "agent-secret",
	Short:   "A secure local secret vault for developers",
	Long: `agent-secret is a secure, developer-friendly CLI tool that acts as a Local Secret Vault.
It eliminates the need to store sensitive credentials in plain-text .env files.

Secrets are stored securely in your OS's native keychain and can be injected
into project-specific .env files when needed.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
