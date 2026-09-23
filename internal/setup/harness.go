package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Harness struct {
	Name    string
	Kind    string
	Detect  func(home, pathEnv string) bool
	Apply   func(home string, command string) (string, error)
	Snippet func(command string) string
}

func Harnesses() []Harness {
	return []Harness{
		{
			Name: "claude",
			Kind: "cli",
			Detect: func(home, pathEnv string) bool {
				return hasBinary(pathEnv, "claude")
			},
			Apply:   applyClaude,
			Snippet: func(command string) string { return "claude mcp add hush -- " + command + " serve" },
		},
		{
			Name: "opencode",
			Kind: "json",
			Detect: func(home, pathEnv string) bool {
				return hasBinary(pathEnv, "opencode") || exists(opencodePath(home))
			},
			Apply: applyJSON(opencodePath, "mcp", func(command string) any {
				return map[string]any{"type": "local", "command": []string{command, "serve"}}
			}),
			Snippet: func(command string) string {
				return "{\"mcp\": {\"hush\": {\"type\": \"local\", \"command\": [\"" + command + "\", \"serve\"]}}}  (en opencode.json)"
			},
		},
		{
			Name: "cursor",
			Kind: "json",
			Detect: func(home, pathEnv string) bool {
				return exists(filepath.Join(home, ".cursor", "mcp.json"))
			},
			Apply: applyJSON(cursorPath, "mcpServers", func(command string) any {
				return map[string]any{"command": command, "args": []string{"serve"}}
			}),
			Snippet: func(command string) string {
				return "{\"mcpServers\": {\"hush\": {\"command\": \"" + command + "\", \"args\": [\"serve\"]}}}  (en ~/.cursor/mcp.json)"
			},
		},
		{
			Name: "codex",
			Kind: "manual",
			Detect: func(home, pathEnv string) bool {
				return hasBinary(pathEnv, "codex")
			},
			Apply: nil,
			Snippet: func(command string) string {
				return "[mcp_servers.hush]\ncommand = \"" + command + "\"\nargs = [\"serve\"]\n# en ~/.codex/config.toml"
			},
		},
	}
}

func hasBinary(pathEnv, name string) bool {
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}
		fi, err := os.Stat(filepath.Join(dir, name))
		if err == nil && !fi.IsDir() {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func opencodePath(home string) string {
	return filepath.Join(home, ".config", "opencode", "opencode.json")
}

func cursorPath(home string) string {
	return filepath.Join(home, ".cursor", "mcp.json")
}

func applyClaude(home string, command string) (string, error) {
	cmd := exec.Command("claude", "mcp", "add", "hush", "--", command, "serve")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("claude mcp add: %s: %w", string(out), err)
	}
	return "registrado vía claude mcp add", nil
}

func applyJSON(pathOf func(string) string, root string, entry func(string) any) func(string, string) (string, error) {
	return func(home string, command string) (string, error) {
		path := pathOf(home)
		var raw []byte
		current, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		var config map[string]any
		if len(current) > 0 {
			if err := json.Unmarshal(current, &config); err != nil {
				return "", fmt.Errorf("%s no es JSON válido, no lo toco: %w", path, err)
			}
			raw = current
		} else {
			config = map[string]any{}
		}
		section, _ := config[root].(map[string]any)
		if section == nil {
			section = map[string]any{}
		}
		if _, ok := section["hush"]; ok {
			return "ya estaba registrado", nil
		}
		section["hush"] = entry(command)
		config[root] = section
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}
		if raw != nil {
			backup := fmt.Sprintf("%s.bak-%d", path, time.Now().Unix())
			if err := os.WriteFile(backup, raw, 0o600); err != nil {
				return "", err
			}
		}
		out, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(path, append(out, '\n'), 0o600); err != nil {
			return "", err
		}
		return "registrado en " + path, nil
	}
}
