package setup

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed skill.md
var skill string

type Target struct {
	Harness string
	Dir     string
}

func Targets(home string) []Target {
	skills := func(parts ...string) string {
		return filepath.Join(append([]string{home}, parts...)...)
	}
	return []Target{
		{"claude", skills(".claude", "skills", "hush")},
		{"codex", skills(".codex", "skills", "hush")},
		{"cursor", skills(".cursor", "skills", "hush")},
		{"agents", skills(".agents", "skills", "hush")},
	}
}

type Result struct {
	Harness string
	Path    string
	State   string
}

func Apply(home string) ([]Result, error) {
	results := []Result{}
	for _, t := range Targets(home) {
		path := filepath.Join(t.Dir, "SKILL.md")
		state, err := writeCurrent(path, []byte(skill))
		if err != nil {
			return nil, err
		}
		results = append(results, Result{Harness: t.Harness, Path: path, State: state})
	}
	return results, nil
}

func writeCurrent(path string, content []byte) (string, error) {
	if current, err := os.ReadFile(path); err == nil && string(current) == string(content) {
		return "al día", nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", err
	}
	return "instalado", nil
}

const wiring = `hush serve es el servidor MCP (stdio). Registralo en tu harness:

  claude mcp add hush -- hush serve

  opencode (opencode.json):
    {"mcp": {"hush": {"type": "local", "command": ["hush", "serve"]}}}

  cursor (~/.cursor/mcp.json):
    {"mcpServers": {"hush": {"command": "hush", "args": ["serve"]}}}

  codex (~/.codex/config.toml):
    [mcp_servers.hush]
    command = "hush"
    args = ["serve"]
`

func Wiring() string {
	return wiring
}
