package cli

import (
	"fmt"

	"envctl/internal/app"
	"envctl/internal/runner"

	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:   "run -- [COMMAND] [args]...",
		Args:  cobra.ArbitraryArgs,
		Short: "Execute command with profile environment injected",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command provided")
			}
			passphrase, err := resolvePassphrase()
			if err != nil {
				return err
			}
			service, err := app.NewVaultService(resolveVaultPath(), passphrase)
			if err != nil {
				return err
			}
			vault, err := service.Load()
			if err != nil {
				return err
			}
			activeProfile, err := resolveProfile(profile, vault)
			if err != nil {
				return err
			}
			vars, ok := vault.ListProfile(activeProfile)
			if !ok {
				return fmt.Errorf("profile not found: %s", activeProfile)
			}
			env := map[string]string{}
			for key, secret := range vars {
				env[key] = secret.Value
			}
			if err := runner.RunCommand(args, env); err != nil {
				if exitErr, ok := err.(*runner.CommandExitError); ok {
					return exitErr
				}
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Profile name (defaults to active profile)")
	return cmd
}
