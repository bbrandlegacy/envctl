package cli

import (
	"bytes"
	"fmt"
	"io"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"envctl/internal/app"
	"envctl/internal/domain"
)

const defaultVaultPath = ".envctl/vault.age"

func resolveVaultPath() string {
	if strings.TrimSpace(cfg.vaultPath) != "" {
		return cfg.vaultPath
	}
	return defaultVaultPath
}

func resolvePassphrase() (string, error) {
	if strings.TrimSpace(cfg.passphraseFile) != "" {
		raw, err := os.ReadFile(cfg.passphraseFile)
		if err != nil {
			return "", err
		}
		return normalizePassphrase(string(raw)), nil
	}

	if passphrase := os.Getenv("ENVCTL_PASSPHRASE"); passphrase != "" {
		return passphrase, nil
	}

	return "", fmt.Errorf("missing passphrase: use --passphrase-file or ENVCTL_PASSPHRASE")
}

func normalizePassphrase(value string) string {
	return strings.TrimRight(value, "\r\n")
}

func resolveProfile(profile string, vault *domain.Vault) (string, error) {
	if strings.TrimSpace(profile) != "" {
		return profile, nil
	}
	if strings.TrimSpace(vault.ActiveProfile) != "" {
		return vault.ActiveProfile, nil
	}
	return "", fmt.Errorf("no active profile: create or select one with envctl profile create/use")
}

func absPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return path
	}
	path = filepath.Clean(path)
	return path
}

func withVault(write bool, fn func(*app.VaultService, *domain.Vault) error) error {
	passphrase, err := resolvePassphrase()
	if err != nil {
		return err
	}

	vaultPath := absPath(resolveVaultPath())
	service, err := app.NewVaultService(vaultPath, passphrase)
	if err != nil {
		return err
	}

	vault, err := service.Load()
	if err != nil {
		return err
	}

	if err := fn(service, vault); err != nil {
		return err
	}

	if write {
		if err := service.Save(vault); err != nil {
			return err
		}
	}

	return nil
}

func writeJSON(out io.Writer, value any) error {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return err
	}
	_, err := io.Copy(out, buffer)
	return err
}
