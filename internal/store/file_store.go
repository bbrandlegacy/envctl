package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"envctl/internal/crypto"
	"envctl/internal/domain"
)

type VaultStore struct {
	path string
}

func NewVaultStore(path string) *VaultStore {
	return &VaultStore{path: filepath.Clean(path)}
}

func (s *VaultStore) Path() string {
	return s.path
}

func (s *VaultStore) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

func (s *VaultStore) EnsureDir() error {
	dir := filepath.Dir(s.path)
	return os.MkdirAll(dir, 0o700)
}

func (s *VaultStore) Load(passphrase string) (*domain.Vault, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("vault does not exist at %s", s.path)
		}
		return nil, err
	}
	decrypted, err := crypto.Decrypt(data, passphrase)
	if err != nil {
		return nil, fmt.Errorf("invalid vault passphrase or corrupt vault")
	}
	return domain.ParseVault(decrypted)
}

func (s *VaultStore) Save(v *domain.Vault, passphrase string) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}
	if v.Version == 0 {
		v.Version = domain.CurrentVaultVersion
	}
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	encrypted, err := crypto.Encrypt(payload, passphrase)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, encrypted, 0o600)
}
