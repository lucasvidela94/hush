package cli

import (
	"bytes"
	"io"
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

func TestSetAceptaMultilinea(t *testing.T) {
	store := testStore(t, nil)
	r, w, _ := os.Pipe()
	_, _ = w.WriteString("uno\ndos")
	_ = w.Close()
	var out bytes.Buffer
	if code := Run([]string{"set", "M"}, store, r, &out, &out); code != OK {
		t.Fatalf("exit %d", code)
	}
	values, _ := store.Load()
	if values["M"] != "uno\ndos" {
		t.Fatalf("roundtrip roto: %q", values["M"])
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

func TestExportMuestraValores(t *testing.T) {
	store := testStore(t, map[string]string{"B": "dos", "A": "uno"})
	var out, errB bytes.Buffer
	allow := func(io.Writer) bool { return true }
	if code := exportSecrets([]string{"A", "B"}, store, &out, &errB, allow); code != OK {
		t.Fatalf("exit %d", code)
	}
	got := out.String()
	if !strings.Contains(got, "A=uno\n") || !strings.Contains(got, "B=dos\n") {
		t.Fatalf("incompleto: %q", got)
	}
}

func TestExportNombreAusente(t *testing.T) {
	store := testStore(t, map[string]string{"A": "uno"})
	var out, errB bytes.Buffer
	allow := func(io.Writer) bool { return true }
	if code := exportSecrets([]string{"A", "ZZZ"}, store, &out, &errB, allow); code != Missing {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out.String(), "A=uno\n") {
		t.Fatalf("no imprimió el presente: %q", out.String())
	}
}

func TestComandoDesconocidoAvisa(t *testing.T) {
	store := testStore(t, nil)
	var out, errB bytes.Buffer
	if code := Run([]string{"frobnicate"}, store, nil, &out, &errB); code != Usage {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(errB.String(), "frobnicate") {
		t.Fatalf("no nombra el comando: %q", errB.String())
	}
}

func TestCheckSinNombresEsUso(t *testing.T) {
	store := testStore(t, nil)
	var out, errB bytes.Buffer
	if code := Run([]string{"check"}, store, nil, &out, &errB); code != Usage {
		t.Fatalf("exit %d", code)
	}
}

func TestRunOnlySinValorEsUso(t *testing.T) {
	store := testStore(t, nil)
	var out, errB bytes.Buffer
	if code := Run([]string{"run", "--only"}, store, nil, &out, &errB); code != Usage {
		t.Fatalf("exit %d", code)
	}
}

func TestRunSinOnlyNiAllEsUso(t *testing.T) {
	store := testStore(t, map[string]string{"A": "uno"})
	var out, errB bytes.Buffer
	if code := Run([]string{"run", "--", "echo", "hola"}, store, nil, &out, &errB); code != Usage {
		t.Fatalf("exit %d, inyectó todo por defecto", code)
	}
}

func TestHelpPorComando(t *testing.T) {
	store := testStore(t, nil)
	var out bytes.Buffer
	if code := Run([]string{"help", "run"}, store, nil, &out, &out); code != OK {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out.String(), "hush run") {
		t.Fatalf("ayuda vacía: %q", out.String())
	}
}
