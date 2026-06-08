package cli

import (
	"fmt"
	"sort"

	"envctl/internal/app"
	"envctl/internal/domain"
	"envctl/internal/output"

	"github.com/spf13/cobra"
)

func newProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage profiles",
	}

	cmd.AddCommand(newProfileCreateCommand())
	cmd.AddCommand(newProfileListCommand())
	cmd.AddCommand(newProfileUseCommand())
	cmd.AddCommand(newProfileDeleteCommand())

	return cmd
}

func newProfileCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Args:  cobra.ExactArgs(1),
		Short: "Create a new profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(true, func(svc *app.VaultService, vault *domain.Vault) error {
				if err := svc.CreateProfile(vault, args[0]); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "created profile %s\n", args[0])
				return nil
			})
		},
	}
}

func newProfileListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(false, func(_ *app.VaultService, vault *domain.Vault) error {
				names := vault.ProfileNames()
				sort.Strings(names)
				rows := make([][]string, 0, len(names))
				for _, name := range names {
					active := ""
					if vault.ActiveProfile == name {
						active = "*"
					}
					rows = append(rows, []string{name, active})
				}
				output.PrintTable(cmd.OutOrStdout(), []string{"PROFILE", "ACTIVE"}, rows)
				return nil
			})
		},
	}
}

func newProfileUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use [name]",
		Args:  cobra.ExactArgs(1),
		Short: "Set active profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(true, func(svc *app.VaultService, vault *domain.Vault) error {
				if _, ok := vault.Profiles[args[0]]; !ok {
					return fmt.Errorf("profile not found: %s", args[0])
				}
				svc.SetActiveProfile(vault, args[0])
				fmt.Fprintf(cmd.OutOrStdout(), "active profile set to %s\n", args[0])
				return nil
			})
		},
	}
}

func newProfileDeleteCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete [name]",
		Args:  cobra.ExactArgs(1),
		Short: "Delete a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(true, func(svc *app.VaultService, vault *domain.Vault) error {
				if err := svc.DeleteProfile(vault, args[0], force); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "deleted profile %s\n", args[0])
				return nil
			})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Delete even when profile contains variables")
	return cmd
}
