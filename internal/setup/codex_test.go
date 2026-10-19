package setup

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"hush/internal/vault"
)

func TestProtectCodex(t *testing.T) {
	home := t.TempDir()
	userRules := filepath.Join(home, ".codex", "rules", "default.rules")
	writeConfig(t, userRules, `prefix_rule(pattern = ["hush", "run"], decision = "allow")`+"\n")
	h := harnessNamed(t, "codex")
	if h.Protected(home) {
		t.Fatal("protegido antes de aplicar")
	}
	if _, err := h.Protect(home); err != nil {
		t.Fatal(err)
	}
	if !h.Protected(home) {
		t.Fatal("doctor no lo ve protegido")
	}
	if msg, err := h.Protect(home); err != nil || !strings.Contains(msg, "ya estaba") {
		t.Fatalf("reaplicar: %q %v", msg, err)
	}
	if raw, _ := os.ReadFile(userRules); !strings.Contains(string(raw), `decision = "allow"`) {
		t.Fatalf("tocó las reglas del usuario: %q", raw)
	}

	codex, err := exec.LookPath("codex")
	if err != nil {
		t.Log("sin codex en PATH: no valido las reglas con execpolicy")
		return
	}
	cases := []struct {
		command []string
		want    string
	}{
		{[]string{"hush", "run", "--only", "X", "--", "ls"}, "prompt"},
		{[]string{"hush", "stdin", "X", "--", "cat"}, "prompt"},
		{[]string{"hush", "export"}, "prompt"},
		{[]string{"hush", "check", "X"}, ""},
	}
	for _, tc := range cases {
		args := []string{"execpolicy", "check", "--rules", userRules, "--rules", codexRulesPath(home)}
		out, err := exec.CommandContext(t.Context(), codex, append(args, tc.command...)...).Output()
		if err != nil {
			t.Fatalf("%v: %v", tc.command, err)
		}
		var result struct {
			Decision string `json:"decision"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("%s: %v", out, err)
		}
		if result.Decision != tc.want {
			t.Fatalf("%v: decisión %q, quiero %q", tc.command, result.Decision, tc.want)
		}
	}
}

func TestDoctorSandboxCodex(t *testing.T) {
	cases := []struct {
		name   string
		config string
		want   string
	}{
		{"sin codex", "", ""},
		{"sin deny", "model = \"gpt\"\n", "− sandbox codex"},
		{"con deny", "default_permissions = \"hush\"\n\n[permissions.hush.filesystem]\n\"~/.hush\" = \"deny\"\n", "✓ sandbox codex"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeOpencode(t, "")
			home := t.TempDir()
			if tc.config != "" {
				writeConfig(t, filepath.Join(home, ".codex", "config.toml"), tc.config)
			}
			var out bytes.Buffer
			Doctor(home, vault.New(filepath.Join(home, ".hush")), "v9.9.9", "/bin/hush", &out)
			if tc.want == "" {
				if strings.Contains(out.String(), "sandbox codex") {
					t.Fatalf("avisa sandbox sin codex:\n%s", out.String())
				}
				return
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Fatalf("quiero %q en:\n%s", tc.want, out.String())
			}
		})
	}
}
