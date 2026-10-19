package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func harnessNamed(t *testing.T, name string) Harness {
	t.Helper()
	for _, h := range Harnesses() {
		if h.Name == name {
			return h
		}
	}
	t.Fatalf("sin harness %s", name)
	return Harness{}
}

func writeConfig(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readConfig(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatalf("%s: %v", raw, err)
	}
	return config
}

func stringList(list any) []string {
	out := []string{}
	for _, v := range list.([]any) {
		out = append(out, v.(string))
	}
	return out
}

func TestProtectClaude(t *testing.T) {
	cases := []struct {
		name      string
		original  string
		wantDeny  []string
		wantAllow []string
		wantModel bool
	}{
		{"sin settings", "", []string{"Read(~/.hush/**)", "Edit(~/.hush/**)"}, nil, false},
		{
			"fusiona sin pisar",
			`{"model": "opus", "permissions": {"allow": ["Bash(ls *)"], "deny": ["Read(./.env)"]}}`,
			[]string{"Read(./.env)", "Read(~/.hush/**)", "Edit(~/.hush/**)"},
			[]string{"Bash(ls *)"},
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			path := filepath.Join(home, ".claude", "settings.json")
			if tc.original != "" {
				writeConfig(t, path, tc.original)
			}
			h := harnessNamed(t, "claude")
			if h.Protected(home) {
				t.Fatal("protegido antes de aplicar")
			}
			if _, err := h.Protect(home); err != nil {
				t.Fatal(err)
			}
			config := readConfig(t, path)
			perms := config["permissions"].(map[string]any)
			if got := stringList(perms["deny"]); strings.Join(got, ",") != strings.Join(tc.wantDeny, ",") {
				t.Fatalf("deny %v, quiero %v", got, tc.wantDeny)
			}
			if got := stringList(perms["ask"]); strings.Join(got, ",") != "Bash(hush run *),Bash(hush stdin *),Bash(hush export *)" {
				t.Fatalf("ask %v", got)
			}
			if tc.wantAllow != nil && strings.Join(stringList(perms["allow"]), ",") != strings.Join(tc.wantAllow, ",") {
				t.Fatalf("allow pisado: %v", perms["allow"])
			}
			if tc.wantModel && config["model"] != "opus" {
				t.Fatalf("perdió model: %v", config)
			}
			if !h.Protected(home) {
				t.Fatal("doctor no lo ve protegido")
			}
			if msg, err := h.Protect(home); err != nil || !strings.Contains(msg, "ya estaba") {
				t.Fatalf("reaplicar: %q %v", msg, err)
			}
		})
	}
}

func TestProtectOpencodeV2AgregaAlFinal(t *testing.T) {
	home := t.TempDir()
	path := opencodePath(home)
	writeConfig(t, path, `{"mcp": {"servers": {}}, "permissions": [{"action": "shell", "resource": "*", "effect": "allow"}]}`)
	h := harnessNamed(t, "opencode")
	if h.Protected(home) {
		t.Fatal("protegido antes de aplicar")
	}
	if _, err := h.Protect(home); err != nil {
		t.Fatal(err)
	}
	rules := readConfig(t, path)["permissions"].([]any)
	first := rules[0].(map[string]any)
	if first["resource"] != "*" || first["effect"] != "allow" {
		t.Fatalf("movió la regla del usuario: %v", rules)
	}
	last := map[string]string{}
	for _, r := range rules[1:] {
		m := r.(map[string]any)
		last[m["action"].(string)+" "+m["resource"].(string)] = m["effect"].(string)
	}
	want := map[string]string{
		"external_directory ~/.hush/*": "deny",
		"read ~/.hush/*":               "deny",
		"edit ~/.hush/*":               "deny",
		"shell *.hush/*":               "deny",
		"shell hush run *":             "ask",
		"shell hush stdin *":           "ask",
		"shell hush export *":          "ask",
	}
	for k, v := range want {
		if last[k] != v {
			t.Fatalf("falta %s=%s en %v", k, v, rules)
		}
	}
	if !h.Protected(home) {
		t.Fatal("doctor no lo ve protegido")
	}
	if msg, err := h.Protect(home); err != nil || !strings.Contains(msg, "ya estaba") {
		t.Fatalf("reaplicar: %q %v", msg, err)
	}
}

func TestProtectNoTocaConfigDudosa(t *testing.T) {
	cases := []struct {
		name     string
		harness  string
		path     func(string) string
		original string
	}{
		{"claude JSON inválido", "claude", claudeSettingsPath, `{"model": `},
		{"claude permissions no es objeto", "claude", claudeSettingsPath, `{"permissions": []}`},
		{"opencode v1", "opencode", opencodePath, `{"mcp": {"x": {}}, "permission": {"bash": "allow"}}`},
		{"opencode sin marca de versión", "opencode", opencodePath, `{"model": "x"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			path := tc.path(home)
			writeConfig(t, path, tc.original)
			if _, err := harnessNamed(t, tc.harness).Protect(home); err == nil {
				t.Fatal("aplicó sobre una config que no entiende")
			}
			raw, _ := os.ReadFile(path)
			if string(raw) != tc.original {
				t.Fatalf("modificó %q → %q", tc.original, raw)
			}
		})
	}
}
