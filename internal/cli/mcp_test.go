package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func mcpParams(t *testing.T, value any) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	return payload
}

func TestMCPInitializeAndToolsList(t *testing.T) {
	mcpAllowExec = false
	t.Cleanup(func() { mcpAllowExec = false })

	result, err := handleMCPRequest(mcpRequest{JSONRPC: "2.0", ID: float64(1), Method: "initialize"})
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	initResult := result.(map[string]interface{})
	if initResult["name"] != "envctl" {
		t.Fatalf("unexpected initialize result: %#v", initResult)
	}

	result, err = handleMCPRequest(mcpRequest{JSONRPC: "2.0", ID: float64(2), Method: "tools/list"})
	if err != nil {
		t.Fatalf("tools/list failed: %v", err)
	}
	text := stringifyMCPResult(t, result)
	if !strings.Contains(text, "envctl_context") || !strings.Contains(text, "envctl_exec") {
		t.Fatalf("expected context and exec tools in list: %s", text)
	}
	if !strings.Contains(text, "disabled") || !strings.Contains(text, "profile") || !strings.Contains(text, "command") {
		t.Fatalf("expected disabled marker and useful schemas: %s", text)
	}
}

func TestMCPContextSafeAndExecGated(t *testing.T) {
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
		mcpAllowExec = false
	})
	t.Setenv("ENVCTL_PASSPHRASE", "mcp-passphrase")

	secretValue := "SMOKE_MCP_SECRET_VALUE_79ad"
	mustRunEnvctlForTest(t, "init")
	mustRunEnvctlForTest(t, "profile", "create", "dev")
	mustRunEnvctlForTest(t, "profile", "use", "dev")
	mustRunEnvctlForTest(t, "secrets", "set", "API_TOKEN", secretValue)
	if err := os.WriteFile(".envdesc", []byte("API_TOKEN: token - API bearer token\n"), 0o600); err != nil {
		t.Fatalf("write .envdesc: %v", err)
	}

	mcpAllowExec = false
	contextResult, err := handleMCPRequest(mcpRequest{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params: mcpParams(t, mcpToolCall{
			Name:      "envctl_context",
			Arguments: map[string]interface{}{"profile": "dev", "envdesc": ".envdesc"},
		}),
	})
	if err != nil {
		t.Fatalf("envctl_context failed: %v", err)
	}
	contextText := stringifyMCPResult(t, contextResult)
	if strings.Contains(contextText, secretValue) {
		t.Fatalf("mcp context leaked raw secret: %s", contextText)
	}
	if !strings.Contains(contextText, "API_TOKEN") || !strings.Contains(contextText, "SET") {
		t.Fatalf("mcp context missing safe key/status: %s", contextText)
	}

	_, err = handleMCPRequest(mcpRequest{
		JSONRPC: "2.0",
		ID:      float64(2),
		Method:  "tools/call",
		Params: mcpParams(t, mcpToolCall{
			Name:      "envctl_exec",
			Arguments: map[string]interface{}{"profile": "dev", "command": []string{"sh", "-c", "printf disabled"}},
		}),
	})
	if err == nil || !strings.Contains(err.Error(), "disabled by default") {
		t.Fatalf("expected exec disabled error, got %v", err)
	}

	mcpAllowExec = true
	execResult, err := handleMCPRequest(mcpRequest{
		JSONRPC: "2.0",
		ID:      float64(3),
		Method:  "tools/call",
		Params: mcpParams(t, mcpToolCall{
			Name:      "envctl_exec",
			Arguments: map[string]interface{}{"profile": "dev", "command": []string{"sh", "-c", `test "$API_TOKEN" = "$1" && printf MCP_EXEC_OK`, "sh", secretValue}},
		}),
	})
	if err != nil {
		t.Fatalf("exec enabled failed: %v", err)
	}
	execText := stringifyMCPResult(t, execResult)
	if !strings.Contains(execText, "MCP_EXEC_OK") {
		t.Fatalf("expected exec success output, got: %s", execText)
	}
}

func stringifyMCPResult(t *testing.T, result any) string {
	t.Helper()
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	return string(payload)
}
