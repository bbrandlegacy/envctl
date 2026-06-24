package store

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"envctl/internal/domain"
)

func TestVaultStoreSaveLoadAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".envctl", "vault.age")
	store := NewVaultStore(path)

	vault := domain.NewVault()
	if err := vault.CreateProfile("dev"); err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := vault.SetSecret("dev", "API_TOKEN", "STORE_SECRET_VALUE"); err != nil {
		t.Fatalf("set secret: %v", err)
	}

	if err := store.Save(vault, "passphrase"); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	loaded, err := store.Load("passphrase")
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}
	value, ok, err := loaded.GetSecret("dev", "API_TOKEN")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if !ok || value != "STORE_SECRET_VALUE" {
		t.Fatalf("unexpected loaded secret ok=%t value=%q", ok, value)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat vault: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("vault permissions = %o, want 600", got)
		}
		dirInfo, err := os.Stat(filepath.Dir(path))
		if err != nil {
			t.Fatalf("stat vault dir: %v", err)
		}
		if got := dirInfo.Mode().Perm(); got != 0o700 {
			t.Fatalf("vault dir permissions = %o, want 700", got)
		}
	}
}

func TestVaultStoreEnsureDirTightensExistingDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions not enforced on windows")
	}
	path := filepath.Join(t.TempDir(), ".envctl", "vault.age")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create lax dir: %v", err)
	}
	store := NewVaultStore(path)
	if err := store.Save(domain.NewVault(), "passphrase"); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("vault dir permissions = %o, want tightened 700", got)
	}
}

func TestVaultStoreAtomicSaveCleansTemporaryFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".envctl", "vault.age")
	store := NewVaultStore(path)
	vault := domain.NewVault()

	if err := store.Save(vault, "passphrase"); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".vault-*.tmp"))
	if err != nil {
		t.Fatalf("glob temp files: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no leftover temp files, found %v", matches)
	}

	vault.ActiveProfile = "dev"
	if err := store.Save(vault, "passphrase"); err != nil {
		t.Fatalf("second save vault: %v", err)
	}
	loaded, err := store.Load("passphrase")
	if err != nil {
		t.Fatalf("load after second save: %v", err)
	}
	if loaded.ActiveProfile != "dev" {
		t.Fatalf("expected replacement save to persist active profile, got %q", loaded.ActiveProfile)
	}
}
