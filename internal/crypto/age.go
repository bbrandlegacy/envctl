package crypto

import (
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
)

func Encrypt(data []byte, passphrase string) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return nil, fmt.Errorf("age recipient: %w", err)
	}
	var output bytes.Buffer
	writer, err := age.Encrypt(&output, recipient)
	if err != nil {
		return nil, fmt.Errorf("age encrypt init: %w", err)
	}
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("age encrypt write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("age encrypt close: %w", err)
	}
	return output.Bytes(), nil
}

func Decrypt(data []byte, passphrase string) ([]byte, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, fmt.Errorf("age identity: %w", err)
	}
	reader := bytes.NewReader(data)
	decrypted, err := age.Decrypt(reader, identity)
	if err != nil {
		return nil, fmt.Errorf("age decrypt: %w", err)
	}
	decryptedPayload, err := io.ReadAll(decrypted)
	if err != nil {
		return nil, fmt.Errorf("age read decrypted: %w", err)
	}
	return decryptedPayload, nil
}

