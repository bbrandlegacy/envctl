package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runEnvctlForTest(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runEnvctlWithInputForTest(t, nil, args...)
}

func runEnvctlWithInputForTest(t *testing.T, input *strings.Reader, args ...string) (string, error) {
	t.Helper()
	cfg = commandConfig{}
	cmd := NewRootCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if input != nil {
		cmd.SetIn(input)
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func mustRunEnvctlForTest(t *testing.T, args ...string) string {
	t.Helper()
	out, err := runEnvctlForTest(t, args...)
	if err != nil {
		t.Fatalf("envctl %s failed: %v\noutput:\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

func assertDoesNotContain(t *testing.T, output string, forbidden ...string) {
	t.Helper()
	for _, value := range forbidden {
		if value == "" {
			continue
		}
		if strings.Contains(output, value) {
			t.Fatalf("output leaked forbidden value %q:\n%s", value, output)
		}
	}
}

func TestCLIIntegrationSmokeAndRedaction(t *testing.T) {
	tmp := t.TempDir()
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWD)
	})
	t.Setenv("ENVCTL_PASSPHRASE", "smoke-passphrase")

	devToken := "SMOKE_SECRET_DEV_TOKEN_9aef4c6b"
	prodToken := "SMOKE_SECRET_PROD_TOKEN_184f27de"
	devDB := "SMOKE_SECRET_DEV_DB_VALUE_51d37b"
	prodDB := "SMOKE_SECRET_PROD_DB_VALUE_671df2"
	forbidden := []string{devToken, prodToken, devDB, prodDB}

	mustRunEnvctlForTest(t, "init")
	mustRunEnvctlForTest(t, "profile", "create", "dev")
	mustRunEnvctlForTest(t, "profile", "create", "prod")
	mustRunEnvctlForTest(t, "profile", "use", "dev")
	mustRunEnvctlForTest(t, "secrets", "set", "API_TOKEN", devToken, "--profile", "dev")
	mustRunEnvctlForTest(t, "secrets", "set", "DATABASE_URL", devDB, "--profile", "dev")
	mustRunEnvctlForTest(t, "secrets", "set", "API_TOKEN", prodToken, "--profile", "prod")
	mustRunEnvctlForTest(t, "secrets", "set", "DATABASE_URL", prodDB, "--profile", "prod")

	if err := os.WriteFile(".envdesc", []byte("API_TOKEN: token - API bearer token\nDATABASE_URL: url - Primary database URL\n"), 0o600); err != nil {
		t.Fatalf("write .envdesc: %v", err)
	}

	listOut := mustRunEnvctlForTest(t, "secrets", "list", "--profile", "dev")
	assertDoesNotContain(t, listOut, forbidden...)
	if !strings.Contains(listOut, "API_TOKEN") || !strings.Contains(listOut, "***") {
		t.Fatalf("expected masked key list, got:\n%s", listOut)
	}

	listJSONOut := mustRunEnvctlForTest(t, "secrets", "list", "--profile", "dev", "--json")
	assertDoesNotContain(t, listJSONOut, forbidden...)

	contextOut := mustRunEnvctlForTest(t, "context", "--profile", "dev", "--envdesc", ".envdesc")
	assertDoesNotContain(t, contextOut, forbidden...)
	if !strings.Contains(contextOut, "SET") || !strings.Contains(contextOut, "API_TOKEN") {
		t.Fatalf("expected safe context entries, got:\n%s", contextOut)
	}

	contextJSONOut := mustRunEnvctlForTest(t, "context", "--profile", "dev", "--envdesc", ".envdesc", "--json")
	assertDoesNotContain(t, contextJSONOut, forbidden...)

	diffOut := mustRunEnvctlForTest(t, "diff", "dev", "prod")
	assertDoesNotContain(t, diffOut, forbidden...)
	if !strings.Contains(diffOut, "CHANGED") || !strings.Contains(diffOut, "API_TOKEN") {
		t.Fatalf("expected changed keys in diff, got:\n%s", diffOut)
	}

	diffJSONOut := mustRunEnvctlForTest(t, "diff", "dev", "prod", "--json")
	assertDoesNotContain(t, diffJSONOut, forbidden...)

	getOut := mustRunEnvctlForTest(t, "secrets", "get", "API_TOKEN", "--profile", "dev")
	if strings.TrimSpace(getOut) != devToken {
		t.Fatalf("expected explicit get to return raw token %q, got %q", devToken, strings.TrimSpace(getOut))
	}

	runMarker := filepath.Join(tmp, "run-ok.txt")
	mustRunEnvctlForTest(t, "run", "--profile", "dev", "--", "sh", "-c", `test "$API_TOKEN" = "$1" && printf RUN_ENV_OK > "$2"`, "sh", devToken, runMarker)
	marker, err := os.ReadFile(runMarker)
	if err != nil {
		t.Fatalf("expected run marker file: %v", err)
	}
	if string(marker) != "RUN_ENV_OK" {
		t.Fatalf("unexpected run marker: %q", string(marker))
	}
}

func TestSecretsSetFromStdinAndInputValidation(t *testing.T) {
	tmp := t.TempDir()
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWD)
	})
	t.Setenv("ENVCTL_PASSPHRASE", "stdin-passphrase")

	stdinValue := "SMOKE_STDIN_SECRET_VALUE_2f4a7"
	mustRunEnvctlForTest(t, "init")
	mustRunEnvctlForTest(t, "profile", "create", "dev")
	mustRunEnvctlForTest(t, "profile", "use", "dev")

	out, err := runEnvctlWithInputForTest(t, strings.NewReader(stdinValue+"\n"), "secrets", "set", "API_TOKEN", "--stdin")
	if err != nil {
		t.Fatalf("set --stdin failed: %v\noutput:\n%s", err, out)
	}
	getOut := mustRunEnvctlForTest(t, "secrets", "get", "API_TOKEN")
	if strings.TrimSpace(getOut) != stdinValue {
		t.Fatalf("expected stdin value %q, got %q", stdinValue, strings.TrimSpace(getOut))
	}

	_, err = runEnvctlWithInputForTest(t, strings.NewReader("other\n"), "secrets", "set", "API_TOKEN", "positional", "--stdin")
	if err == nil || !strings.Contains(err.Error(), "only one mode") {
		t.Fatalf("expected conflicting input mode error, got %v", err)
	}

	_, err = runEnvctlForTest(t, "secrets", "set", "API_TOKEN")
	if err == nil || !strings.Contains(err.Error(), "secret value required") {
		t.Fatalf("expected missing value error, got %v", err)
	}

	_, err = runEnvctlForTest(t, "secrets", "set", "API_TOKEN", "--prompt")
	if err == nil || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("expected non-terminal prompt error, got %v", err)
	}
}
