package runner

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRedactaEnv(t *testing.T) {
	var out, errB bytes.Buffer
	code := Run(
		map[string]string{"HUSH_T": "valor-ultrasecreto"},
		[]string{"sh", "-c", "echo la key es $HUSH_T"},
		nil, &out, &errB,
	)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(out.String(), "valor-ultrasecreto") {
		t.Fatalf("salida sin redactar: %q", out.String())
	}
	if !strings.Contains(out.String(), "[REDACTED]") {
		t.Fatalf("falta marca: %q", out.String())
	}
}

func TestPipeRedactaEco(t *testing.T) {
	var out, errB bytes.Buffer
	code := Pipe("valor-ultrasecreto", []string{"cat"}, &out, &errB)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(out.String(), "valor-ultrasecreto") {
		t.Fatalf("salida sin redactar: %q", out.String())
	}
}

func TestRunPropagaExit(t *testing.T) {
	var out, errB bytes.Buffer
	code := Run(nil, []string{"sh", "-c", "exit 7"}, nil, &out, &errB)
	if code != 7 {
		t.Fatalf("exit %d, quiero 7", code)
	}
}
