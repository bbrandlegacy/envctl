package cli

import (
	"fmt"
	"sort"
	"strings"

	"envctl/internal/app"
	"envctl/internal/domain"
	"envctl/internal/output"

	"github.com/spf13/cobra"
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
	return &cobra.Command{
		Use:   "set [KEY] [VALUE]",
		Args:  cobra.ExactArgs(2),
		Short: "Set or update a secret in a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(true, func(svc *app.VaultService, vault *domain.Vault) error {
				activeProfile, err := resolveProfile(profile, vault)
				if err != nil {
					return err
				}
				if err := svc.SetSecret(vault, activeProfile, args[0], args[1]); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "set %s in profile %s\n", args[0], activeProfile)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Profile name (defaults to active profile)")
	return cmd
}

func newGetCommand() *cobra.Command {
	var profile string
	return &cobra.Command{
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
	return &cobra.Command{
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
	return &cobra.Command{
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
