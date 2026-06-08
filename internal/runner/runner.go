package runner

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type CommandExitError struct {
	Code    int
	Message string
}

func (e *CommandExitError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("command failed with exit code %d", e.Code)
}

func RunCommand(command []string, env map[string]string) error {
	if len(command) == 0 {
		return fmt.Errorf("command not provided")
	}

	baseEnv := os.Environ()
	envMap := map[string]string{}
	for _, entry := range baseEnv {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envMap[parts[0]] = parts[1]
	}

	for key, value := range env {
		envMap[key] = value
	}

	merged := make([]string, 0, len(envMap))
	for key, value := range envMap {
		merged = append(merged, key+"="+value)
	}
	sort.Strings(merged)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = merged
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &CommandExitError{Code: exitErr.ExitCode(), Message: "child process failed"}
		}
		return err
	}
	return nil
}
