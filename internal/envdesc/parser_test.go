package envdesc

import (
	"os"
	"strings"
	"testing"
)

func TestParseEnvdesc(t *testing.T) {
	content := strings.Join([]string{
		"# comment",
		"",
		"DATABASE_URL: url - PostgreSQL connection string",
		"FEATURE_FLAG?: bool - Enables feature X",
	}, "\n")

	path := t.TempDir() + "/envdesc"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture failed: %v", err)
	}

	meta, err := Parse(path)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got := meta["DATABASE_URL"]; got.Type != "url" || got.Description == "" {
		t.Fatalf("unexpected metadata for DATABASE_URL: %+v", got)
	}
	if got := meta["FEATURE_FLAG"]; !got.Optional {
		t.Fatalf("expected optional flag")
	}
}

func TestParseEnvdescErrors(t *testing.T) {
	path := t.TempDir() + "/envdesc"
	bad := []byte("BAD LINE WITHOUT COLON")
	if err := os.WriteFile(path, bad, 0o600); err != nil {
		t.Fatalf("write fixture failed: %v", err)
	}
	if _, err := Parse(path); err == nil {
		t.Fatalf("expected parse error")
	}
}
