package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	bad := []map[string]string{
		{"": "x"},
		{"A=B": "x"},
		{"A\nB": "x"},
		{"A#B": "x"},
		{" A": "x"},
	}
	for _, values := range bad {
		if err := Validate(values); err == nil {
			t.Fatalf("aceptó %v", values)
		}
	}
	if err := Validate(map[string]string{"A_B-9.x": "ok\nmultilínea con \\ backslash"}); err != nil {
		t.Fatalf("rechazó válido: %v", err)
	}
}

func TestEscapeRoundtrip(t *testing.T) {
	store := New(t.TempDir())
	want := map[string]string{
		"PEM": "-----BEGIN-----\nabc\\def\r\n-----END-----",
		"URL": "a=b&c=d",
	}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("clave %s: got %q want %q", k, got[k], v)
		}
	}
}

func TestSaveRechazaCorrupto(t *testing.T) {
	store := New(t.TempDir())
	if err := store.Save(map[string]string{"A=B": "x"}); err == nil {
		t.Fatal("guardó nombre con =")
	}
}

func TestLoadLegacySinUnescape(t *testing.T) {
	dir := t.TempDir()
	legacy := "VIEJA=C:\\new\\temp\nOTRA=ab\\ncd\\\\e\n"
	if err := os.WriteFile(filepath.Join(dir, "vault"), []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := New(dir).Load()
	if err != nil {
		t.Fatal(err)
	}
	if got["VIEJA"] != `C:\new\temp` || got["OTRA"] != `ab\ncd\\e` {
		t.Fatalf("migró mal: %q", got)
	}
}

func TestLockExclusivo(t *testing.T) {
	dir := t.TempDir()
	l1, err := Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer l1.Release()
	done := make(chan error, 1)
	go func() {
		l2, err := Acquire(dir)
		if err == nil {
			l2.Release()
		}
		done <- err
	}()
	select {
	case err := <-done:
		t.Fatalf("lock no excluyó: %v", err)
	case <-time.After(200 * time.Millisecond):
	}
}
