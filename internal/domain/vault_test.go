package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestVaultProfileAndSecretLifecycle(t *testing.T) {
	v := NewVault()

	if err := v.CreateProfile("dev"); err != nil {
		t.Fatalf("create profile failed: %v", err)
	}
	if err := v.CreateProfile("dev"); err == nil {
		t.Fatalf("expected duplicate profile error")
	}

	v.SetActiveProfile("dev")
	if v.ActiveProfile != "dev" {
		t.Fatalf("active profile should be set")
	}

	if err := v.SetSecret("dev", "API_TOKEN", "abc"); err != nil {
		t.Fatalf("set secret failed: %v", err)
	}

	value, ok, err := v.GetSecret("dev", "API_TOKEN")
	if err != nil {
		t.Fatalf("get secret failed: %v", err)
	}
	if !ok || value != "abc" {
		t.Fatalf("unexpected secret value: ok=%t value=%q", ok, value)
	}

	vars, ok := v.ListProfile("dev")
	if !ok {
		t.Fatalf("profile should exist")
	}
	secret, ok := vars["API_TOKEN"]
	if !ok {
		t.Fatalf("secret should exist in list")
	}
	parsed, err := time.Parse(time.RFC3339, secret.UpdatedAt)
	if err != nil {
		t.Fatalf("updatedAt must be RFC3339: %v", err)
	}
	if parsed.IsZero() {
		t.Fatalf("updatedAt must be set")
	}

	removed, err := v.UnsetSecret("dev", "API_TOKEN")
	if err != nil {
		t.Fatalf("unset secret failed: %v", err)
	}
	if !removed {
		t.Fatalf("expected secret removed")
	}
	_, ok = v.ListProfile("dev")
	if !ok {
		t.Fatalf("profile should still exist")
	}
	_, ok = v.ListProfile("missing")
	if ok {
		t.Fatalf("missing profile should return false")
	}
}

func TestVaultParseValidation(t *testing.T) {
	if _, err := ParseVault([]byte(`{"profiles":{}}`)); err == nil {
		t.Fatalf("expected missing version error")
	}

	v := NewVault()
	v.Version = CurrentVaultVersion
	v.Profiles["dev"] = Profile{}
	payload, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	loaded, err := ParseVault(payload)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if loaded.Version != CurrentVaultVersion {
		t.Fatalf("version mismatch")
	}
}
