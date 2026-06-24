package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type IntegrationTarget string

const (
	TargetGeneric  IntegrationTarget = "generic"
	TargetClaude   IntegrationTarget = "claude"
	TargetChatGPT  IntegrationTarget = "chatgpt"
	TargetCursor   IntegrationTarget = "cursor"
	TargetOpenAIFn IntegrationTarget = "openai-functions"
)

type SkillCommand struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Command     string   `json:"command"`
	Tags        []string `json:"tags"`
}

type SkillManifest struct {
	Schema         string         `json:"schema"`
	Name           string         `json:"name"`
	Version        int            `json:"version"`
	Target         string         `json:"target"`
	DefaultProfile string         `json:"defaultProfile"`
	VaultPath      string         `json:"vaultPath"`
	Binary         string         `json:"binary"`
	Commands       []SkillCommand `json:"commands"`
	Notes          []string       `json:"notes"`
}

var supportedTargets = []IntegrationTarget{
	TargetGeneric,
	TargetClaude,
	TargetChatGPT,
	TargetCursor,
	TargetOpenAIFn,
}

func SupportedTargets() []string {
	values := make([]string, 0, len(supportedTargets))
	for _, target := range supportedTargets {
		values = append(values, string(target))
	}
	return values
}

func ValidateTarget(target string) bool {
	for _, item := range supportedTargets {
		if string(item) == target {
			return true
		}
	}
	return false
}

func DefaultManifestPath(target IntegrationTarget, global bool) (string, error) {
	if !ValidateTarget(string(target)) {
		return "", fmt.Errorf("unsupported target: %s", target)
	}
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, ".config", "envctl", "skills", string(target)+".skill.json"), nil
	}
	return filepath.Join(".envctl", "skills", string(target)+".skill.json"), nil
}

func RenderManifest(target IntegrationTarget, defaultProfile, vaultPath, binary string, includeExec, includeSensitiveGet bool) ([]byte, error) {
	if !ValidateTarget(string(target)) {
		return nil, fmt.Errorf("unsupported target: %s", target)
	}
	commands := []SkillCommand{
		{
			Name:        "envctl_context",
			Description: "Inspect environment shape and variable metadata without leaking values.",
			Command:     binary + " context --profile ${profile} --json",
			Tags:        []string{"safe", "context", "metadata"},
		},
	}
	notes := []string{
		"Safe outputs such as context/list/diff expose names, metadata, and SET/MISSING state only; they do not expose raw values.",
	}
	if includeExec {
		commands = append(commands, SkillCommand{
			Name:        "envctl_run",
			Description: "Execute a command with injected profile variables. Treat child output as potentially sensitive.",
			Command:     binary + " run --profile ${profile} -- ${command}",
			Tags:        []string{"runtime", "secrets", "execution", "sensitive-output"},
		})
		notes = append(notes, "Exec manifest mode is enabled: command execution may return child process output; treat that output as potentially sensitive if the child prints environment values.")
	} else {
		notes = append(notes, "Command execution is omitted by default. Regenerate with the exec option only for trusted local workflows that need runtime injection.")
	}
	if includeSensitiveGet {
		commands = append(commands, SkillCommand{
			Name:        "envctl_get",
			Description: "Retrieve a single raw value when explicitly needed. This exposes sensitive data.",
			Command:     binary + " secrets get ${key} --profile ${profile}",
			Tags:        []string{"explicit", "sensitive", "raw-secret"},
		})
		notes = append(notes, "Privileged manifest mode is enabled: envctl_get can expose raw secret values and should only be granted to trusted local operators.")
	} else {
		notes = append(notes, "Raw secret retrieval is omitted by default. Regenerate with the privileged/sensitive option only when a trusted workflow explicitly needs raw secret access.")
	}
	manifest := SkillManifest{
		Schema:         "envctl-skill/v1",
		Name:           "envctl",
		Version:        1,
		Target:         string(target),
		DefaultProfile: defaultProfile,
		VaultPath:      vaultPath,
		Binary:         binary,
		Commands:       commands,
		Notes:          notes,
	}

	return json.MarshalIndent(manifest, "", "  ")
}

func WriteManifest(path string, data []byte, force bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("manifest already exists: %s", path)
		}
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}
