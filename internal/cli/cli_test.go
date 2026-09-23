package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"hush/internal/vault"
)

func testStore(t *testing.T, values map[string]string) vault.Store {
	t.Helper()
	store := vault.New(t.TempDir())
	if err := store.Save(values); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestSetRechazaNombreInvalido(t *testing.T) {
	store := testStore(t, nil)
	r, w, _ := os.Pipe()
	_, _ = w.WriteString("x")
	_ = w.Close()
	var out bytes.Buffer
	code := Run([]string{"set", "A=B"}, store, r, &out, &out)
	if code == OK {
		t.Fatal("aceptó nombre con =")
	}
}

func TestSetRechazaMultilinea(t *testing.T) {
	store := testStore(t, nil)
	r, w, _ := os.Pipe()
	_, _ = w.WriteString("uno\ndos")
	_ = w.Close()
	var out bytes.Buffer
	if code := Run([]string{"set", "M"}, store, r, &out, &out); code == OK {
		t.Fatal("aceptó valor multilínea")
	}
	values, _ := store.Load()
	if len(values) != 0 {
		t.Fatalf("vault corrupto: %v", values)
	}
}

func TestCheckSoloNombres(t *testing.T) {
	store := testStore(t, map[string]string{"A": "valor-a-largo"})
	var out, errB bytes.Buffer
	if code := Run([]string{"check", "A", "B"}, store, nil, &out, &errB); code != Missing {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(out.String(), "valor-a-largo") || strings.Contains(errB.String(), "valor-a-largo") {
		t.Fatal("valor visible en check")
	}
}

func TestExportRechazaSinTTY(t *testing.T) {
	store := testStore(t, map[string]string{"A": "valor-a-largo"})
	var out, errB bytes.Buffer
	if code := Run([]string{"export", "A"}, store, nil, &out, &errB); code == OK {
		t.Fatal("export permitió salida capturada")
	}
	if strings.Contains(out.String(), "valor-a-largo") {
		t.Fatal("valor visible en export no-TTY")
	}
}

func TestCanRevealRechazaBuffer(t *testing.T) {
	if canReveal(&bytes.Buffer{}) {
		t.Fatal("buffer no es terminal")
	}
}
