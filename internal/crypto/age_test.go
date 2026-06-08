package crypto

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	passphrase := "unit-passphrase"
	original := []byte(`{"version":1}`)

	encrypted, err := Encrypt(original, passphrase)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if len(encrypted) == 0 {
		t.Fatalf("encrypted payload should not be empty")
	}

	decrypted, err := Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if string(decrypted) != string(original) {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	passphrase := "unit-passphrase"
	wrong := "wrong-passphrase"

	encrypted, err := Encrypt([]byte("secret"), passphrase)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if _, err := Decrypt(encrypted, wrong); err == nil {
		t.Fatalf("expected decrypt error with wrong passphrase")
	}
}
