package cli

import (
	"fmt"
	"strings"

	"envctl/internal/ai"

	"github.com/spf13/cobra"
)

func newAICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "AI integration tooling",
	}

	cmd.AddCommand(newAIInstallSkillCommand())
	return cmd
}

func newAIInstallSkillCommand() *cobra.Command {
	var target string
	var global bool
	var outputPath string
	var apply bool
	var force bool
	var defaultProfile string

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Generate and install an envctl skill manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			target = strings.TrimSpace(strings.ToLower(target))
			if target == "" {
				target = string(ai.TargetGeneric)
			}
			if !ai.ValidateTarget(target) {
				return fmt.Errorf("unsupported AI target '%s' (use one of: %s)", target, strings.Join(ai.SupportedTargets(), ", "))
			}

			payload, err := ai.RenderManifest(ai.IntegrationTarget(target), defaultProfile, resolveVaultPath(), "envctl")
			if err != nil {
				return err
			}

			autoPath, autoPathErr := ai.DefaultManifestPath(ai.IntegrationTarget(target), global)

			if !apply {
				pathHint := "<unavailable: cannot determine global path>"
				if autoPathErr == nil {
					pathHint = autoPath
				}
				_, _ = fmt.Fprintf(
					cmd.OutOrStdout(),
					"%s\n\n# Preview mode. Use --apply to write this manifest to disk.\n# Suggested path: %s\n",
					string(payload),
					pathHint,
				)
				return nil
			}

			if outputPath == "" {
				if autoPathErr != nil {
					return autoPathErr
				}
				outputPath = autoPath
			}

			if err := ai.WriteManifest(outputPath, payload, force); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "installed skill manifest: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", string(ai.TargetGeneric), "Skill target format (generic, claude, chatgpt, cursor, openai-functions)")
	cmd.Flags().StringVarP(&outputPath, "path", "p", "", "Output file path (defaults to global or local auto path)")
	cmd.Flags().BoolVar(&global, "global", true, "Write under global config directory when path is not provided")
	cmd.Flags().BoolVar(&apply, "apply", false, "Write manifest to disk")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing manifest file")
	cmd.Flags().StringVar(&defaultProfile, "default-profile", "", "Default profile in generated manifest")

	return cmd
}
