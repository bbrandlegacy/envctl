package cli

import (
	"fmt"
	"sort"
	"strings"

	"envctl/internal/app"
	"envctl/internal/domain"
	"envctl/internal/envdesc"
	"envctl/internal/output"

	"github.com/spf13/cobra"
)

func newContextCommand() *cobra.Command {
	var profile string
	var asJSON bool
	var envDescPath string

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Generate AI-safe profile context",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVault(false, func(_ *app.VaultService, vault *domain.Vault) error {
				activeProfile, err := resolveProfile(profile, vault)
				if err != nil {
					return err
				}
				secrets, ok := vault.ListProfile(activeProfile)
				if !ok {
					return fmt.Errorf("profile not found: %s", activeProfile)
				}
				metadata, err := envdesc.Parse(envDescPath)
				if err != nil {
					return err
				}
				keys := map[string]struct{}{}
				for key := range secrets {
					keys[key] = struct{}{}
				}
				for key := range metadata {
					keys[key] = struct{}{}
				}
				allKeys := make([]string, 0, len(keys))
				for key := range keys {
					allKeys = append(allKeys, key)
				}
				sort.Strings(allKeys)

				type contextEntry struct {
					Key         string `json:"key"`
					Status      string `json:"status"`
					Type        string `json:"type"`
					Description string `json:"description"`
					Optional    bool   `json:"optional"`
				}

				rows := make([][]string, 0, len(allKeys))
				payload := make([]contextEntry, 0, len(allKeys))
				for _, key := range allKeys {
					meta, hasMeta := metadata[key]
					_, hasValue := secrets[key]
					status := "MISSING"
					if hasValue {
						status = "SET"
					}
					typ := inferType(key, hasValue, secrets)
					description := ""
					optional := false
					if hasMeta {
						typ = meta.Type
						description = meta.Description
						optional = meta.Optional
					}
					if strings.TrimSpace(description) == "" {
						description = "No metadata available"
					}
					rows = append(rows, []string{key, status, typ, description, fmt.Sprintf("%t", optional)})
					payload = append(payload, contextEntry{Key: key, Status: status, Type: typ, Description: description, Optional: optional})
				}
				if asJSON {
					return writeJSON(cmd.OutOrStdout(), payload)
				}
				output.PrintTable(cmd.OutOrStdout(), []string{"KEY", "STATUS", "TYPE", "DESCRIPTION", "OPTIONAL"}, rows)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Profile name (defaults to active profile)")
	cmd.Flags().StringVarP(&envDescPath, "envdesc", "e", ".envdesc", "Path to .envdesc metadata")
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "Output JSON")
	return cmd
}

func inferType(key string, hasValue bool, secrets map[string]domain.Secret) string {
	upper := strings.ToUpper(key)
	if strings.Contains(upper, "URL") {
		return "url"
	}
	if strings.HasSuffix(upper, "_PORT") {
		return "int"
	}
	if strings.Contains(upper, "JSON") {
		return "json"
	}
	if strings.HasSuffix(upper, "_FLAG") || strings.HasSuffix(upper, "_ENABLED") {
		return "bool"
	}
	if strings.HasSuffix(upper, "_PATH") || strings.HasSuffix(upper, "_FILE") {
		return "path"
	}
	if hasValue {
		if strings.Contains(upper, "SECRET") || strings.Contains(upper, "TOKEN") || strings.Contains(upper, "KEY") {
			return "string"
		}
	}
	return "string"
}
