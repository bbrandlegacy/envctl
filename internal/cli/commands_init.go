package cli

import (
	"fmt"

	"envctl/internal/app"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize an encrypted vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			passphrase, err := resolvePassphrase()
			if err != nil {
				return err
			}
			service, err := app.NewVaultService(resolveVaultPath(), passphrase)
			if err != nil {
				return err
			}
			if _, err := service.Init(force); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "vault initialized at %s\n", resolveVaultPath())
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing vault")
	return cmd
}
