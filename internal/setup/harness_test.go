package setup

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectaPorBinario(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "opencode"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	seen := map[string]bool{}
	for _, h := range Harnesses() {
		if h.Detect(home, dir) {
			seen[h.Name] = true
		}
	}
	if !seen["opencode"] {
		t.Fatal("no detectó opencode en PATH")
	}
	if seen["claude"] {
		t.Fatal("falso positivo claude")
	}
}

func TestApplyJSONFusionaYRespalda(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "opencode.json")
	original := `{"mcp": {"otro": {"type": "local", "command": ["x"]}}}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	var apply func(string, string) (string, error)
	for _, h := range Harnesses() {
		if h.Name == "opencode" {
			apply = h.Apply
		}
	}
	msg, err := apply(home, "/bin/hush")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg, path) {
		t.Fatalf("mensaje: %q", msg)
	}
	raw, _ := os.ReadFile(path)
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	mcp := config["mcp"].(map[string]any)
	if mcp["otro"] == nil || mcp["hush"] == nil {
		t.Fatalf("fusión rota: %s", raw)
	}
	matches, _ := filepath.Glob(path + ".bak-*")
	if len(matches) != 1 {
		t.Fatal("sin backup")
	}
	if msg2, err := apply(home, "/bin/hush"); err != nil || !strings.Contains(msg2, "ya estaba") {
		t.Fatalf("reaplicar: %q %v", msg2, err)
	}
}

func TestApplyJSONRechazaInvalido(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "mcp.json")
	if err := os.WriteFile(path, []byte("{roto"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, h := range Harnesses() {
		if h.Name == "cursor" {
			if _, err := h.Apply(home, "/bin/hush"); err == nil {
				t.Fatal("tocó JSON inválido")
			}
		}
	}
	if got, _ := os.ReadFile(path); string(got) != "{roto" {
		t.Fatal("modificó archivo inválido")
	}
}

func TestCodexEsManual(t *testing.T) {
	for _, h := range Harnesses() {
		if h.Name == "codex" && h.Apply != nil {
			t.Fatal("codex debería ser manual (TOML)")
		}
	}
}

func TestWizardSinHarness(t *testing.T) {
	home := t.TempDir()
	r, w, _ := os.Pipe()
	_ = w.Close()
	var out bytes.Buffer
	t.Setenv("PATH", t.TempDir())
	if code := Wizard(home, r, &out, "/bin/hush"); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out.String(), "No detecté") {
		t.Fatalf("falta mensaje: %q", out.String())
	}
}
