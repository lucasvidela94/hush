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
