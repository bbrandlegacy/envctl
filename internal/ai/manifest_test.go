package ai

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderManifest(t *testing.T) {
	manifest, err := RenderManifest(TargetGeneric, "dev", ".envctl/vault.age", "envctl")
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(string(manifest), "\"target\": \"generic\"") {
		t.Fatalf("expected generic target")
	}
	if !strings.Contains(string(manifest), "envctl run --profile ${profile} -- ${command}") {
		t.Fatalf("expected run command")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(manifest, &payload); err != nil {
		t.Fatalf("valid JSON expected: %v", err)
	}
}

func TestSupportedTargetsContainsExpectedValues(t *testing.T) {
	targets := SupportedTargets()
	if len(targets) == 0 {
		t.Fatalf("expected targets")
	}
	expected := map[string]bool{
		string(TargetGeneric):   false,
		string(TargetClaude):    false,
		string(TargetChatGPT):   false,
		string(TargetCursor):    false,
		string(TargetOpenAIFn):  false,
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
