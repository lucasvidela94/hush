package runner

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const persistedSecret = "valor-ultrasecreto"

func runIn(t *testing.T, script string) (home, cwd, stderr string) {
	t.Helper()
	home, cwd = t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Chdir(cwd)
	var out, errB bytes.Buffer
	Run(context.Background(), map[string]string{"HUSH_T": persistedSecret}, []string{"sh", "-c", script}, nil, &out, &errB)
	return home, cwd, errB.String() + out.String()
}

func TestAvisaSiElHijoGuardaElValor(t *testing.T) {
	cases := []struct {
		name   string
		script string
		path   func(home, cwd string) string
	}{
		{"claude.json", `printf '{"k":"%s"}' "$HUSH_T" > "$HOME/.claude.json"`, func(h, _ string) string { return filepath.Join(h, ".claude.json") }},
		{"env local", `echo "KEY=$HUSH_T" > .env.local`, func(_, c string) string { return filepath.Join(c, ".env.local") }},
		{"mcp.json", `echo "$HUSH_T" > .mcp.json`, func(_, c string) string { return filepath.Join(c, ".mcp.json") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home, cwd, got := runIn(t, tc.script)
			if !strings.Contains(got, "HUSH_T") || !strings.Contains(got, tc.path(home, cwd)) {
				t.Fatalf("sin aviso con nombre y ruta: %q", got)
			}
			if strings.Contains(got, persistedSecret) {
				t.Fatalf("el aviso muestra el valor: %q", got)
			}
		})
	}
}

func TestNoAvisaSinEscritura(t *testing.T) {
	_, _, got := runIn(t, `echo hola`)
	if strings.Contains(got, "aviso") {
		t.Fatalf("aviso innecesario: %q", got)
	}
}

func TestNoAvisaSiElArchivoNoTieneElValor(t *testing.T) {
	_, _, got := runIn(t, `echo "KEY=otra-cosa" > .env`)
	if strings.Contains(got, "aviso") {
		t.Fatalf("aviso innecesario: %q", got)
	}
}

func TestNoAvisaPorArchivoPreexistenteSinCambios(t *testing.T) {
	home, cwd := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Chdir(cwd)
	if err := os.WriteFile(filepath.Join(cwd, ".env"), []byte("KEY="+persistedSecret), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errB bytes.Buffer
	Run(context.Background(), map[string]string{"HUSH_T": persistedSecret}, []string{"sh", "-c", "echo hola"}, nil, &out, &errB)
	if strings.Contains(errB.String(), "aviso") {
		t.Fatalf("avisó por un archivo que el hijo no tocó: %q", errB.String())
	}
}

func TestNoLeeArchivosFueraDeLaLista(t *testing.T) {
	_, _, got := runIn(t, `echo "$HUSH_T" > notas.txt`)
	if strings.Contains(got, "aviso") {
		t.Fatalf("avisó por un archivo fuera de la lista: %q", got)
	}
}
