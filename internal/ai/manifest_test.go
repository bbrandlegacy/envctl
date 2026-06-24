package ai

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderManifestSafeByDefault(t *testing.T) {
	manifest, err := RenderManifest(TargetGeneric, "dev", ".envctl/vault.age", "envctl", false, false)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	payloadText := string(manifest)
	if !strings.Contains(payloadText, "\"target\": \"generic\"") {
		t.Fatalf("expected generic target")
	}
	if !strings.Contains(payloadText, "envctl context --profile ${profile} --json") {
		t.Fatalf("expected safe context command")
	}
	if strings.Contains(payloadText, "envctl_get") || strings.Contains(payloadText, "secrets get") || strings.Contains(payloadText, "envctl_run") || strings.Contains(payloadText, " run --profile") {
		t.Fatalf("safe default manifest must not include raw get or command execution:\n%s", payloadText)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(manifest, &payload); err != nil {
		t.Fatalf("valid JSON expected: %v", err)
	}
}

func TestRenderManifestPrivilegedIncludesSensitiveGet(t *testing.T) {
	manifest, err := RenderManifest(TargetGeneric, "dev", ".envctl/vault.age", "envctl", false, true)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	payloadText := string(manifest)
	if !strings.Contains(payloadText, "envctl_get") || !strings.Contains(payloadText, "secrets get") {
		t.Fatalf("privileged manifest should include raw get command:\n%s", payloadText)
	}
	if !strings.Contains(payloadText, "raw secret") && !strings.Contains(payloadText, "sensitive") {
		t.Fatalf("privileged manifest should warn about sensitive output:\n%s", payloadText)
	}
}

func TestRenderManifestExecIncludesRun(t *testing.T) {
	manifest, err := RenderManifest(TargetGeneric, "dev", ".envctl/vault.age", "envctl", true, false)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	payloadText := string(manifest)
	if !strings.Contains(payloadText, "envctl_run") || !strings.Contains(payloadText, "run --profile") {
		t.Fatalf("exec manifest should include run command:\n%s", payloadText)
	}
	if strings.Contains(payloadText, "envctl_get") || strings.Contains(payloadText, "secrets get") {
		t.Fatalf("exec-only manifest should not include raw get command:\n%s", payloadText)
	}
}

func TestSupportedTargetsContainsExpectedValues(t *testing.T) {
	targets := SupportedTargets()
	if len(targets) == 0 {
		t.Fatalf("expected targets")
	}
	expected := map[string]bool{
		string(TargetGeneric):  false,
		string(TargetClaude):   false,
		string(TargetChatGPT):  false,
		string(TargetCursor):   false,
		string(TargetOpenAIFn): false,
	}
	for _, target := range targets {
		if _, ok := expected[target]; ok {
			expected[target] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Fatalf("missing target in SupportedTargets: %s", name)
		}
	}
}
