package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand builds the envctl command tree.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "envctl",
		Short: "Encrypted environment management with AI-safe workflows",
		Long:  "envctl manages local, profile-based secrets encrypted with age and provides AI-safe context output and command execution.",
	}

	rootCmd.PersistentFlags().StringVar(&cfg.vaultPath, "vault", "", "Path to the age-encrypted vault file (default .envctl/vault.age)")
	rootCmd.PersistentFlags().StringVar(&cfg.passphraseFile, "passphrase-file", "", "Path to a file containing the vault passphrase")

	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newProfileCommand())
	rootCmd.AddCommand(newSecretCommand())
	rootCmd.AddCommand(newContextCommand())
	rootCmd.AddCommand(newDiffCommand())
	rootCmd.AddCommand(newRunCommand())
	rootCmd.AddCommand(newAICommand())
	rootCmd.AddCommand(newMCPCommand())

	return rootCmd
}

var cfg commandConfig

type commandConfig struct {
	vaultPath      string
	passphraseFile string
}
