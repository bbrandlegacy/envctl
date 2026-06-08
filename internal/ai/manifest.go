package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type IntegrationTarget string

const (
	TargetGeneric   IntegrationTarget = "generic"
	TargetClaude    IntegrationTarget = "claude"
	TargetChatGPT   IntegrationTarget = "chatgpt"
	TargetCursor    IntegrationTarget = "cursor"
	TargetOpenAIFn  IntegrationTarget = "openai-functions"
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

func RenderManifest(target IntegrationTarget, defaultProfile, vaultPath, binary string) ([]byte, error) {
	if !ValidateTarget(string(target)) {
		return nil, fmt.Errorf("unsupported target: %s", target)
	}
	manifest := SkillManifest{
		Schema:         "envctl-skill/v1",
		Name:           "envctl",
		Version:        1,
		Target:         string(target),
		DefaultProfile: defaultProfile,
		VaultPath:      vaultPath,
		Binary:         binary,
		Commands: []SkillCommand{
			{
				Name:        "envctl_context",
				Description: "Inspect environment shape and variable metadata without leaking values.",
				Command:     binary + " context --profile ${profile} --json",
				Tags:        []string{"safe", "context", "metadata"},
			},
			{
				Name:        "envctl_run",
				Description: "Execute a command with injected profile variables.",
				Command:     binary + " run --profile ${profile} -- ${command}",
				Tags:        []string{"runtime", "secrets", "execution"},
			},
			{
				Name:        "envctl_get",
				Description: "Retrieve a single raw value when explicitly needed.",
				Command:     binary + " secrets get ${key} --profile ${profile}",
				Tags:        []string{"explicit", "sensitive"},
			},
		},
		Notes: []string{
			"Use only the explicit commands and keep raw values out of prompts unless using envctl get.",
			"This manifest is transport-agnostic; wire it to Claude, ChatGPT, Cursor, or your internal assistant runtime as needed.",
		},
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
