package envdesc

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Meta struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Optional    bool   `json:"optional"`
}

func Parse(path string) (map[string]Meta, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]Meta{}, nil
		}
		return nil, err
	}
	defer file.Close()

	result := map[string]Meta{}
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: expected KEY: type - description", lineNum)
		}
		rawKey := strings.TrimSpace(parts[0])
		entryTypeAndDescription := strings.TrimSpace(parts[1])

		if rawKey == "" || entryTypeAndDescription == "" {
			return nil, fmt.Errorf("line %d: malformed .envdesc entry", lineNum)
		}

		entry := Meta{}
		entry.Optional = strings.HasSuffix(rawKey, "?")
		entry.Key = strings.TrimSuffix(rawKey, "?")
		if entry.Key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}
		nameValue := strings.SplitN(entryTypeAndDescription, "-", 2)
		if len(nameValue) != 2 {
			return nil, fmt.Errorf("line %d: expected type - description", lineNum)
		}
		entry.Type = strings.TrimSpace(nameValue[0])
		entry.Description = strings.TrimSpace(nameValue[1])
		if entry.Type == "" {
			return nil, fmt.Errorf("line %d: missing type", lineNum)
		}
		if entry.Description == "" {
			return nil, fmt.Errorf("line %d: missing description", lineNum)
		}

		result[entry.Key] = entry
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
