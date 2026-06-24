package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"envctl/internal/app"
	"envctl/internal/domain"
	"envctl/internal/envdesc"

	"github.com/spf13/cobra"
)

func newMCPCommand() *cobra.Command {
	var transport string
	var allowExec bool

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start experimental MCP-style stdio adapter",
		RunE: func(cmd *cobra.Command, args []string) error {
			if transport != "stdio" {
				return fmt.Errorf("unsupported transport: %s", transport)
			}
			mcpAllowExec = allowExec || os.Getenv("ENVCTL_MCP_ALLOW_EXEC") == "1"
			return serveMCP()
		},
	}

	cmd.Flags().StringVarP(&transport, "transport", "t", "stdio", "Transport mode (currently stdio)")
	cmd.Flags().BoolVar(&allowExec, "allow-exec", false, "Enable envctl_exec tool for command execution with injected secrets")
	return cmd
}

var mcpAllowExec bool

func serveMCP() error {
	in := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriterSize(os.Stdout, 4096)
	defer out.Flush()

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}

		var req mcpRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			_ = writeMCPError(out, nil, "invalid_request", fmt.Sprintf("invalid json: %v", err))
			continue
		}

		result, err := handleMCPRequest(req)
		if err != nil {
			code := "request_failed"
			if mcpErr, ok := err.(*mcpRequestError); ok {
				code = mcpErr.Code
			}
			if req.ID == nil {
				continue
			}
			_ = writeMCPError(out, req.ID, code, err.Error())
			continue
		}

		resp := mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}
		payload, marshalErr := json.Marshal(resp)
		if marshalErr != nil {
			_ = writeMCPError(out, req.ID, "internal_error", marshalErr.Error())
			continue
		}
		_, _ = out.Write(append(payload, '\n'))
		if flushErr := out.Flush(); flushErr != nil {
			return flushErr
		}
	}

	if err := in.Err(); err != nil {
		return err
	}
	return nil
}

func handleMCPRequest(req mcpRequest) (interface{}, error) {
	if req.JSONRPC != "" && req.JSONRPC != "2.0" {
		return nil, &mcpRequestError{Code: "invalid_request", Message: "jsonrpc must be 2.0"}
	}
	if req.Method == "" {
		return nil, &mcpRequestError{Code: "invalid_request", Message: "missing method"}
	}

	switch req.Method {
	case "initialize":
		var params map[string]interface{}
		if len(req.Params) > 0 && string(req.Params) != "null" {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return nil, &mcpRequestError{Code: "invalid_params", Message: fmt.Sprintf("invalid initialize params: %v", err)}
			}
		}
		return map[string]interface{}{
			"name":     "envctl",
			"version":  "1.0",
			"protocol": "mcp-like",
			"params":   params,
		}, nil
	case "tools/list":
		tools := []map[string]interface{}{
			{
				"name":        "envctl_context",
				"description": "Build AI-safe environment context for a profile without raw values.",
				"inputSchema": map[string]interface{}{
					"type":     "object",
					"required": []string{},
					"properties": map[string]interface{}{
						"profile": map[string]interface{}{"type": "string", "description": "Profile name; defaults to active profile."},
						"envdesc": map[string]interface{}{"type": "string", "description": "Path to .envdesc metadata file."},
					},
				},
			},
			{
				"name":        "envctl_exec",
				"description": "Execute command with profile-injected secrets. Disabled unless envctl mcp --allow-exec or ENVCTL_MCP_ALLOW_EXEC=1 is set; child output may be sensitive.",
				"disabled":    !mcpAllowExec,
				"inputSchema": map[string]interface{}{
					"type":     "object",
					"required": []string{"command"},
					"properties": map[string]interface{}{
						"profile": map[string]interface{}{"type": "string", "description": "Profile name; defaults to active profile."},
						"command": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "argv vector to execute."},
					},
				},
			},
		}
		return map[string]interface{}{"tools": tools}, nil
	case "tools/call":
		if len(req.Params) == 0 {
			return nil, &mcpRequestError{Code: "invalid_params", Message: "tools/call requires params"}
		}
		var callReq mcpToolCall
		if err := json.Unmarshal(req.Params, &callReq); err != nil {
			return nil, &mcpRequestError{Code: "invalid_params", Message: fmt.Sprintf("invalid tools/call params: %v", err)}
		}
		if strings.TrimSpace(callReq.Name) == "" {
			return nil, &mcpRequestError{Code: "invalid_params", Message: "tools/call requires a non-empty name"}
		}
		switch callReq.Name {
		case "envctl_context":
			args := struct {
				Profile string `json:"profile"`
				EnvDesc string `json:"envdesc"`
			}{Profile: "", EnvDesc: ".envdesc"}
			if err := mapToStruct(callReq.Arguments, &args); err != nil {
				return nil, &mcpRequestError{Code: "invalid_params", Message: fmt.Sprintf("invalid envctl_context args: %v", err)}
			}
			payload, err := buildMCPContext(args.Profile, args.EnvDesc)
			if err != nil {
				return nil, &mcpRequestError{Code: "tool_execution_failed", Message: err.Error()}
			}
			return map[string]interface{}{"profile": args.Profile, "context": payload}, nil
		case "envctl_exec":
			if !mcpAllowExec {
				return nil, &mcpRequestError{Code: "permission_denied", Message: "envctl_exec is disabled by default; restart envctl mcp with --allow-exec or set ENVCTL_MCP_ALLOW_EXEC=1 to enable command execution"}
			}
			args := struct {
				Profile string   `json:"profile"`
				Command []string `json:"command"`
			}{}
			if err := mapToStruct(callReq.Arguments, &args); err != nil {
				return nil, &mcpRequestError{Code: "invalid_params", Message: fmt.Sprintf("invalid envctl_exec args: %v", err)}
			}
			if len(args.Command) == 0 {
				return nil, &mcpRequestError{Code: "invalid_params", Message: "envctl_exec requires 'command'"}
			}
			exitCode, output, err := runCommandForMCP(args.Profile, args.Command)
			if err != nil {
				return map[string]interface{}{
					"command":  args.Command,
					"exitCode": exitCode,
					"output":   output,
					"error":    err.Error(),
				}, nil
			}
			return map[string]interface{}{
				"command":  args.Command,
				"exitCode": exitCode,
				"output":   output,
			}, nil
		default:
			return nil, &mcpRequestError{Code: "method_not_found", Message: fmt.Sprintf("unknown tool: %s", callReq.Name)}
		}
	default:
		return nil, &mcpRequestError{Code: "method_not_found", Message: fmt.Sprintf("unknown method: %s", req.Method)}
	}
}

func buildMCPContext(profile, envdescPath string) (interface{}, error) {
	var payload interface{}
	err := withVault(false, func(_ *app.VaultService, vault *domain.Vault) error {
		activeProfile := profile
		if strings.TrimSpace(activeProfile) == "" {
			activeProfile = vault.ActiveProfile
		}
		if strings.TrimSpace(activeProfile) == "" {
			return fmt.Errorf("no active profile: create or select one with envctl profile create/use")
		}
		secrets, ok := vault.ListProfile(activeProfile)
		if !ok {
			return fmt.Errorf("profile not found: %s", activeProfile)
		}

		metadata, err := envdesc.Parse(envdescPath)
		if err != nil {
			return err
		}

		keys := map[string]struct{}{}
		for key := range secrets {
			keys[key] = struct{}{}
		}
		for key := range metadata {
			keys[key] = struct{}{}
		}
		allKeys := make([]string, 0, len(keys))
		for key := range keys {
			allKeys = append(allKeys, key)
		}
		sort.Strings(allKeys)

		rows := make([]map[string]interface{}, 0, len(allKeys))
		for _, key := range allKeys {
			meta, hasMeta := metadata[key]
			_, hasValue := secrets[key]
			status := "MISSING"
			if hasValue {
				status = "SET"
			}
			typ := inferType(key, hasValue, secrets)
			description := ""
			optional := false
			if hasMeta {
				typ = meta.Type
				description = meta.Description
				optional = meta.Optional
			}
			if strings.TrimSpace(description) == "" {
				description = "No metadata available"
			}
			rows = append(rows, map[string]interface{}{
				"key":         key,
				"status":      status,
				"type":        typ,
				"description": description,
				"optional":    optional,
			})
		}

		payload = map[string]interface{}{
			"profile": activeProfile,
			"vars":    rows,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func runCommandForMCP(profile string, command []string) (int, string, error) {
	exitCode := 1
	var output string

	err := withVault(false, func(_ *app.VaultService, vault *domain.Vault) error {
		activeProfile := profile
		if strings.TrimSpace(activeProfile) == "" {
			activeProfile = vault.ActiveProfile
		}
		if strings.TrimSpace(activeProfile) == "" {
			return fmt.Errorf("no active profile: create or select one with envctl profile create/use")
		}

		vars, ok := vault.ListProfile(activeProfile)
		if !ok {
			return fmt.Errorf("profile not found: %s", activeProfile)
		}
		env := map[string]string{}
		for key, secret := range vars {
			env[key] = secret.Value
		}

		execOutput, execErr := runCommandForMCPInternal(command, env)
		output = execOutput
		if execErr != nil {
			var exit *exec.ExitError
			if errors.As(execErr, &exit) {
				exitCode = exit.ExitCode()
			} else {
				exitCode = 1
			}
			return execErr
		}
		exitCode = 0
		return nil
	})
	if err != nil {
		return exitCode, output, err
	}
	return exitCode, output, nil
}

func runCommandForMCPInternal(command []string, env map[string]string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("command not provided")
	}

	baseEnv := os.Environ()
	envMap := map[string]string{}
	for _, entry := range baseEnv {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	for key, value := range env {
		envMap[key] = value
	}

	envList := make([]string, 0, len(envMap))
	for key, value := range envMap {
		envList = append(envList, key+"="+value)
	}
	sort.Strings(envList)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = envList
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func writeMCPError(out *bufio.Writer, id interface{}, code, message string) error {
	resp := mcpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &mcpError{
			Code:    code,
			Message: message,
		},
	}
	payload, _ := json.Marshal(resp)
	_, err := out.Write(append(payload, '\n'))
	_ = out.Flush()
	return err
}

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type mcpResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *mcpError   `json:"error,omitempty"`
}

type mcpError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type mcpToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type mcpRequestError struct {
	Code    string
	Message string
}

func (e *mcpRequestError) Error() string {
	return e.Message
}

func mapToStruct(value map[string]interface{}, output interface{}) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, output)
}
