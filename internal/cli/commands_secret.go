package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"envctl/internal/app"
	"envctl/internal/domain"
	"envctl/internal/output"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage environment secrets",
	}

	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newUnsetCommand())
	cmd.AddCommand(newListCommand())

	return cmd
}

func newSetCommand() *cobra.Command {
	var profile string
	var fromStdin bool
	var fromPrompt bool
	cmd := &cobra.Command{
		Use:   "set [KEY] [VALUE]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Set or update a secret in a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := resolveSecretSetValue(cmd, args, fromStdin, fromPrompt)
			if err != nil {
				return err
			}
			return withVault(true, func(svc *app.VaultService, vault *domain.Vault) error {
				activeProfile, err := resolveProfile(profile, vault)
				if err != nil {
					return err
				}
				if err := svc.SetSecret(vault, activeProfile, args[0], value); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "set %s in profile %s\n", args[0], activeProfile)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Profile name (defaults to active profile)")
	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "Read the secret value from stdin")
	cmd.Flags().BoolVar(&fromPrompt, "prompt", false, "Prompt for the secret value without echo when possible")
	return cmd
}

func resolveSecretSetValue(cmd *cobra.Command, args []string, fromStdin, fromPrompt bool) (string, error) {
	modeCount := 0
	if len(args) == 2 {
		modeCount++
	}
	if fromStdin {
		modeCount++
	}
	if fromPrompt {
		modeCount++
	}
	if modeCount == 0 {
		return "", fmt.Errorf("secret value required: pass VALUE, --stdin, or --prompt")
	}
	if modeCount > 1 {
		return "", fmt.Errorf("provide the secret value using only one mode: VALUE, --stdin, or --prompt")
	}
	if len(args) == 2 {
		return args[1], nil
	}
	if fromStdin {
		raw, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", err
		}
		return strings.TrimRight(string(raw), "\r\n"), nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("--prompt requires an interactive terminal; use --stdin for piped input")
	}
	_, _ = fmt.Fprint(cmd.ErrOrStderr(), "Secret value: ")
	value, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, _ = fmt.Fprintln(cmd.ErrOrStderr())
	if err != nil {
		return "", fmt.Errorf("read secret prompt: %w", err)
	}
	return string(value), nil
}

func newGetCommand() *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:   "get [KEY]",
		Args:  cobra.ExactArgs(1),
		Short: "Get a raw secret value",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(false, func(_ *app.VaultService, vault *domain.Vault) error {
				activeProfile, err := resolveProfile(profile, vault)
				if err != nil {
					return err
				}
				value, ok, err := vault.GetSecret(activeProfile, args[0])
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("key not found")
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), value)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Profile name (defaults to active profile)")
	return cmd
}

func newUnsetCommand() *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:   "unset [KEY]",
		Args:  cobra.ExactArgs(1),
		Short: "Remove a secret from a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(true, func(svc *app.VaultService, vault *domain.Vault) error {
				activeProfile, err := resolveProfile(profile, vault)
				if err != nil {
					return err
				}
				removed, err := svc.UnsetSecret(vault, activeProfile, args[0])
				if err != nil {
					return err
				}
				if !removed {
					return fmt.Errorf("key not found")
				}
				fmt.Fprintf(cmd.OutOrStdout(), "removed %s from %s\n", args[0], activeProfile)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Profile name (defaults to active profile)")
	return cmd
}

func newListCommand() *cobra.Command {
	var profile string
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets in a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(false, func(_ *app.VaultService, vault *domain.Vault) error {
				activeProfile, err := resolveProfile(profile, vault)
				if err != nil {
					return err
				}
				vars, ok := vault.ListProfile(activeProfile)
				if !ok {
					return fmt.Errorf("profile not found: %s", activeProfile)
				}
				keys := make([]string, 0, len(vars))
				for key := range vars {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				if jsonOutput {
					type masked struct {
						Key       string `json:"key"`
						Value     string `json:"value"`
						UpdatedAt string `json:"updatedAt"`
					}
					payload := make([]masked, 0, len(keys))
					for _, key := range keys {
						payload = append(payload, masked{Key: key, Value: "***", UpdatedAt: vars[key].UpdatedAt})
					}
					return writeJSON(cmd.OutOrStdout(), payload)
				}
				rows := make([][]string, 0, len(keys))
				for _, key := range keys {
					rows = append(rows, []string{key, "***", vars[key].UpdatedAt})
				}
				output.PrintTable(cmd.OutOrStdout(), []string{"KEY", "VALUE", "UPDATED"}, rows)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Profile name (defaults to active profile)")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output JSON")
	return cmd
}

func maskValue(value string) string {
	if value == "" {
		return ""
	}
	return strings.Repeat("*", min(8, len(value)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
