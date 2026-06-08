package app

import (
	"fmt"
	"strings"

	"envctl/internal/domain"
	"envctl/internal/store"
)

// VaultService owns vault-level orchestration.
type VaultService struct {
	store      *store.VaultStore
	passphrase string
}

func NewVaultService(vaultPath, passphrase string) (*VaultService, error) {
	if strings.TrimSpace(passphrase) == "" {
		return nil, fmt.Errorf("vault passphrase cannot be empty")
	}
	return &VaultService{
		store:      store.NewVaultStore(vaultPath),
		passphrase: passphrase,
	}, nil
}

func (s *VaultService) Load() (*domain.Vault, error) {
	return s.store.Load(s.passphrase)
}

func (s *VaultService) Save(v *domain.Vault) error {
	return s.store.Save(v, s.passphrase)
}

func (s *VaultService) Init(force bool) (*domain.Vault, error) {
	if !force && s.store.Exists() {
		return nil, fmt.Errorf("vault already exists at %s. Use --force to overwrite", s.store.Path())
	}
	vault := domain.NewVault()
	return vault, s.store.Save(vault, s.passphrase)
}

func (s *VaultService) CreateProfile(v *domain.Vault, profile string) error {
	return v.CreateProfile(profile)
}

func (s *VaultService) DeleteProfile(v *domain.Vault, profile string, force bool) error {
	return v.DeleteProfile(profile, force)
}

func (s *VaultService) SetSecret(v *domain.Vault, profile, key, value string) error {
	return v.SetSecret(profile, key, value)
}

func (s *VaultService) UnsetSecret(v *domain.Vault, profile, key string) (bool, error) {
	return v.UnsetSecret(profile, key)
}

func (s *VaultService) GetSecret(v *domain.Vault, profile, key string) (string, bool, error) {
	return v.GetSecret(profile, key)
}

func (s *VaultService) ListProfile(v *domain.Vault, profile string) (map[string]domain.Secret, bool) {
	return v.ListProfile(profile)
}

func (s *VaultService) SetActiveProfile(v *domain.Vault, profile string) {
	v.SetActiveProfile(profile)
}
