package cli

import (
	"fmt"
	"sort"

	"envctl/internal/app"
	"envctl/internal/domain"
	"envctl/internal/output"

	"github.com/spf13/cobra"
)

func newDiffCommand() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "diff [PROFILE_A] [PROFILE_B]",
		Args:  cobra.ExactArgs(2),
		Short: "Diff two profiles without exposing values",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(false, func(_ *app.VaultService, vault *domain.Vault) error {
				profileA := args[0]
				profileB := args[1]
				varsA, okA := vault.ListProfile(profileA)
				varsB, okB := vault.ListProfile(profileB)
				if !okA {
					return fmt.Errorf("profile not found: %s", profileA)
				}
				if !okB {
					return fmt.Errorf("profile not found: %s", profileB)
				}

				added := []string{}
				removed := []string{}
				changed := []string{}
				common := []string{}

				seen := map[string]struct{}{}
				for key := range varsA {
					if _, exists := varsB[key]; !exists {
						removed = append(removed, key)
					} else {
						if varsA[key].Value != varsB[key].Value {
							changed = append(changed, key)
						} else {
							common = append(common, key)
						}
					}
					seen[key] = struct{}{}
				}
				for key := range varsB {
					if _, ok := seen[key]; ok {
						continue
					}
					added = append(added, key)
				}

				type diffPayload struct {
					ProfileA string   `json:"profileA"`
					ProfileB string   `json:"profileB"`
					Added    []string `json:"added"`
					Removed  []string `json:"removed"`
					Changed  []string `json:"changed"`
					Common   []string `json:"common"`
				}

				sort.Strings(added)
				sort.Strings(removed)
				sort.Strings(changed)
				sort.Strings(common)

				if asJSON {
					return writeJSON(cmd.OutOrStdout(), diffPayload{
						ProfileA: profileA,
						ProfileB: profileB,
						Added:    added,
						Removed:  removed,
						Changed:  changed,
						Common:   common,
					})
				}

				rows := [][]string{}
				for _, key := range removed {
					rows = append(rows, []string{"REMOVED", key})
				}
				for _, key := range added {
					rows = append(rows, []string{"ADDED", key})
				}
				for _, key := range changed {
					rows = append(rows, []string{"CHANGED", key})
				}
				for _, key := range common {
					rows = append(rows, []string{"SAME", key})
				}

				output.PrintTable(cmd.OutOrStdout(), []string{"STATUS", "KEY"}, rows)
				return nil
			})
		},
	}
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "Output JSON")
	return cmd
}
