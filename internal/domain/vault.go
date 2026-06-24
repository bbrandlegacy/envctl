package domain

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

const CurrentVaultVersion = 1

type Vault struct {
	Version       int                `json:"version"`
	ActiveProfile string             `json:"activeProfile"`
	Profiles      map[string]Profile `json:"profiles"`
}

type Profile struct {
	Vars map[string]Secret `json:"vars"`
}

type Secret struct {
	Value     string `json:"value"`
	UpdatedAt string `json:"updatedAt"`
}

func NewVault() *Vault {
	return &Vault{
		Version:  CurrentVaultVersion,
		Profiles: map[string]Profile{},
	}
}

func ParseVault(data []byte) (*Vault, error) {
	var vault Vault
	if err := json.Unmarshal(data, &vault); err != nil {
		return nil, err
	}

	if vault.Version == 0 {
		return nil, fmt.Errorf("invalid or missing vault version")
	}
	if vault.Version > CurrentVaultVersion {
		return nil, fmt.Errorf("unsupported vault version %d", vault.Version)
	}
	if vault.Profiles == nil {
		vault.Profiles = map[string]Profile{}
	}

	for name, profile := range vault.Profiles {
		if profile.Vars == nil {
			profile.Vars = map[string]Secret{}
			vault.Profiles[name] = profile
		}
	}

	return &vault, nil
}

func (v *Vault) ProfileNames() []string {
	names := make([]string, 0, len(v.Profiles))
	for name := range v.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func validateProfileName(profile string) error {
	if strings.TrimSpace(profile) == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	for _, r := range profile {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			continue
		}
		return fmt.Errorf("invalid profile name: %s", profile)
	}
	return nil
}

func validateEnvKey(key string) error {
	if key == "" {
		return fmt.Errorf("environment key cannot be empty")
	}
	if strings.Contains(key, " ") {
		return fmt.Errorf("environment key cannot contain spaces: %s", key)
	}
	for _, r := range key {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return fmt.Errorf("invalid env key %s", key)
	}
	if (key[0] < 'A' || key[0] > 'Z') && key[0] != '_' {
		return fmt.Errorf("invalid env key %s", key)
	}
	return nil
}

func (v *Vault) ensureProfile(profile string) (Profile, error) {
	if err := validateProfileName(profile); err != nil {
		return Profile{}, err
	}
	p, ok := v.Profiles[profile]
	if !ok {
		p = Profile{Vars: map[string]Secret{}}
		v.Profiles[profile] = p
	}
	if p.Vars == nil {
		p.Vars = map[string]Secret{}
		v.Profiles[profile] = p
	}
	return p, nil
}

func (v *Vault) CreateProfile(profile string) error {
	if _, ok := v.Profiles[profile]; ok {
		return fmt.Errorf("profile already exists: %s", profile)
	}
	_, err := v.ensureProfile(profile)
	return err
}

func (v *Vault) DeleteProfile(profile string, force bool) error {
	if err := validateProfileName(profile); err != nil {
		return err
	}
	p, ok := v.Profiles[profile]
	if !ok {
		return fmt.Errorf("profile not found: %s", profile)
	}
	if len(p.Vars) > 0 && !force {
		return fmt.Errorf("profile %s has secrets; use --force to delete", profile)
	}
	delete(v.Profiles, profile)
	if v.ActiveProfile == profile {
		v.ActiveProfile = ""
	}
	return nil
}

func (v *Vault) SetSecret(profile, key, value string) error {
	if err := validateEnvKey(key); err != nil {
		return err
	}
	p, err := v.ensureProfile(profile)
	if err != nil {
		return err
	}
	if p.Vars == nil {
		p.Vars = map[string]Secret{}
	}
	p.Vars[key] = Secret{
		Value:     value,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	v.Profiles[profile] = p
	return nil
}

func (v *Vault) GetSecret(profile, key string) (string, bool, error) {
	if err := validateEnvKey(key); err != nil {
		return "", false, err
	}
	p, ok := v.Profiles[profile]
	if !ok {
		return "", false, fmt.Errorf("profile not found: %s", profile)
	}
	secret, ok := p.Vars[key]
	if !ok {
		return "", false, nil
	}
	return secret.Value, true, nil
}

func (v *Vault) UnsetSecret(profile, key string) (bool, error) {
	if err := validateEnvKey(key); err != nil {
		return false, err
	}
	p, ok := v.Profiles[profile]
	if !ok {
		return false, fmt.Errorf("profile not found: %s", profile)
	}
	if p.Vars == nil {
		return false, nil
	}
	_, ok = p.Vars[key]
	if !ok {
		return false, nil
	}
	delete(p.Vars, key)
	v.Profiles[profile] = p
	return true, nil
}

func (v *Vault) ListProfile(profile string) (map[string]Secret, bool) {
	p, ok := v.Profiles[profile]
	if !ok {
		return nil, false
	}
	copy := map[string]Secret{}
	for key, value := range p.Vars {
		copy[key] = value
	}
	return copy, true
}

func (v *Vault) SetActiveProfile(profile string) {
	v.ActiveProfile = profile
}
