package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hush/internal/vault"
)

func TestDoctorTodoBien(t *testing.T) {
	home := t.TempDir()
	if _, err := Apply(home); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	entry := `{"mcp": {"hush": {"type": "local", "command": ["hush", "serve"]}}}`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}
	store := vault.New(filepath.Join(home, ".hush"))
	if err := store.Save(map[string]string{"A": "uno"}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if code := Doctor(home, store, "v9.9.9", "/bin/hush", &out); code != 0 {
		t.Fatalf("exit %d:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "v9.9.9") {
		t.Fatalf("sin versión:\n%s", out.String())
	}
}

func TestDoctorReglas(t *testing.T) {
	home := t.TempDir()
	store := vault.New(filepath.Join(home, ".hush"))
	run := func() string {
		var out bytes.Buffer
		Doctor(home, store, "v9.9.9", "/bin/hush", &out)
		return out.String()
	}
	if out := run(); !strings.Contains(out, "− reglas claude: el agente puede leer ~/.hush") {
		t.Fatalf("no avisa reglas faltantes:\n%s", out)
	}
	if _, err := harnessNamed(t, "claude").Protect(home); err != nil {
		t.Fatal(err)
	}
	if out := run(); !strings.Contains(out, "✓ reglas claude") {
		t.Fatalf("no ve las reglas:\n%s", out)
	}
}

func TestDoctorCasaVaciaAvisa(t *testing.T) {
	home := t.TempDir()
	store := vault.New(filepath.Join(home, ".hush"))
	var out bytes.Buffer
	if code := Doctor(home, store, "v9.9.9", "/bin/hush", &out); code != 0 {
		t.Fatalf("casa vacía no es error, es guía: %d", code)
	}
	if !strings.Contains(out.String(), "−") {
		t.Fatalf("sin avisos:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "hush setup") {
		t.Fatalf("sin guía:\n%s", out.String())
	}
}

func TestDoctorDetalleMultilineaNoRompeLaLista(t *testing.T) {
	home := t.TempDir()
	store := vault.New(filepath.Join(home, ".hush"))
	var out bytes.Buffer
	Doctor(home, store, "v9.9.9", "/bin/hush", &out)
	for line := range strings.SplitSeq(strings.TrimRight(out.String(), "\n"), "\n") {
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "      ") &&
			!strings.HasPrefix(line, "  ✓") && !strings.HasPrefix(line, "  −") && !strings.HasPrefix(line, "  ✗") {
			t.Errorf("línea suelta rompe el layout: %q", line)
		}
		if !strings.HasPrefix(line, "  ") {
			t.Errorf("línea sin sangría: %q", line)
		}
	}
	if !strings.Contains(out.String(), "      args = [\"serve\"]") {
		t.Fatalf("el snippet de codex no quedó indentado:\n%s", out.String())
	}
}

func TestDoctorVaultPermisoMalo(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".hush")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "vault"), []byte("A=uno\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := vault.New(dir)
	var out bytes.Buffer
	if code := Doctor(home, store, "v9.9.9", "/bin/hush", &out); code == 0 {
		t.Fatal("permiso 644 dio OK")
	}
}
