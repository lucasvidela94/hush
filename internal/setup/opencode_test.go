package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hush/internal/vault"
)

// fakeOpencode deja en PATH solo un opencode que imprime version; vacío
// significa que no hay opencode instalado.
func fakeOpencode(t *testing.T, version string) {
	t.Helper()
	dir := t.TempDir()
	if version != "" {
		script := "#!/bin/sh\necho '" + version + "'\n"
		if err := os.WriteFile(filepath.Join(dir, "opencode"), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
}

func TestApplyOpencodeSegunVersion(t *testing.T) {
	cases := []struct {
		name      string
		installed string
		original  string
		wantV2    bool
		keep      string
	}{
		{"v2", "", `{"mcp": {"servers": {"exa": {}}}}`, true, "exa"},
		{"v2 con entrada vieja en mcp.hush", "", `{"mcp": {"servers": {}, "hush": {"type": "local"}}}`, true, ""},
		{"v1", "", `{"permission": {}, "mcp": {"otro": {}}}`, false, "otro"},
		{"sin marca con opencode 2", "opencode v2.0.32", `{"model": "x"}`, true, ""},
		{"sin marca sin opencode", "", `{"model": "x"}`, false, ""},
		{"sin archivo con opencode 1", "1.14.2", "", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeOpencode(t, tc.installed)
			home := t.TempDir()
			path := opencodePath(home)
			if tc.original != "" {
				writeConfig(t, path, tc.original)
			}
			h := harnessNamed(t, "opencode")
			if _, err := h.Apply(home, "/bin/hush"); err != nil {
				t.Fatal(err)
			}
			mcp := readConfig(t, path)["mcp"].(map[string]any)
			servers, _ := mcp["servers"].(map[string]any)
			entry := mcp["hush"]
			if tc.wantV2 {
				entry = servers["hush"]
				if _, stale := mcp["hush"]; stale {
					t.Fatalf("quedó mcp.hush en una config v2: %v", mcp)
				}
				if tc.keep != "" && servers[tc.keep] == nil {
					t.Fatalf("perdió %s: %v", tc.keep, mcp)
				}
			} else if tc.keep != "" && mcp[tc.keep] == nil {
				t.Fatalf("perdió %s: %v", tc.keep, mcp)
			}
			command, _ := entry.(map[string]any)["command"].([]any)
			if len(command) != 2 || command[0] != "/bin/hush" || command[1] != "serve" {
				t.Fatalf("entrada mal: %v", mcp)
			}
			if msg, err := h.Apply(home, "/bin/hush"); err != nil || !strings.Contains(msg, "ya estaba") {
				t.Fatalf("reaplicar: %q %v", msg, err)
			}
		})
	}
}

func TestDoctorMCPOpencode(t *testing.T) {
	cases := []struct {
		name     string
		original string
		want     string
	}{
		{"v2 registrado", `{"mcp": {"servers": {"hush": {}}}}`, "✓ mcp opencode"},
		{"v2 con reglas pero sin MCP", `{"mcp": {"servers": {}}, "permissions": [{"action": "read", "resource": "~/.hush/*", "effect": "deny"}]}`, "− mcp opencode: sin entrada hush"},
		{"v2 con entrada v1", `{"mcp": {"servers": {}, "hush": {}}}`, "− mcp opencode: entrada hush en formato v1"},
		{"v1 registrado", `{"permission": {}, "mcp": {"hush": {}}}`, "✓ mcp opencode"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeOpencode(t, "")
			home := t.TempDir()
			writeConfig(t, opencodePath(home), tc.original)
			var out bytes.Buffer
			Doctor(home, vault.New(filepath.Join(home, ".hush")), "v9.9.9", "/bin/hush", &out)
			if !strings.Contains(out.String(), tc.want) {
				t.Fatalf("quiero %q en:\n%s", tc.want, out.String())
			}
		})
	}
}
