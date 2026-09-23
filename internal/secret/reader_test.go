package secret

import (
	"os"
	"testing"
)

func TestReadPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString("  valor-con-espacios  \n"); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	got, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "valor-con-espacios" {
		t.Fatalf("got %q", got)
	}
}

func TestReadPipeVacio(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	got, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q, quiero vacío", got)
	}
}
