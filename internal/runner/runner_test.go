package runner

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunRedactaEnv(t *testing.T) {
	var out, errB bytes.Buffer
	code := Run(
		context.Background(),
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
	code := Pipe(context.Background(), "valor-ultrasecreto", []string{"cat"}, &out, &errB)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(out.String(), "valor-ultrasecreto") {
		t.Fatalf("salida sin redactar: %q", out.String())
	}
}

func TestRunPropagaExit(t *testing.T) {
	var out, errB bytes.Buffer
	code := Run(context.Background(), nil, []string{"sh", "-c", "exit 7"}, nil, &out, &errB)
	if code != 7 {
		t.Fatalf("exit %d, quiero 7", code)
	}
}

func TestMataGrupoCompleto(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	var out, errB bytes.Buffer
	start := time.Now()
	code := Run(ctx, nil, []string{"sh", "-c", "sleep 30 & sleep 30"}, nil, &out, &errB)
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("tardó %v, el grupo sobrevivió", elapsed)
	}
	if code == 0 {
		t.Fatal("exit 0 con kill")
	}
}
