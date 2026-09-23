package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundtrip(t *testing.T) {
	store := New(t.TempDir())
	want := map[string]string{"ALPHA": "uno", "BETA": "dos"}
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

func TestLoadMissingDirVacio(t *testing.T) {
	store := New(filepath.Join(t.TempDir(), "no-existe"))
	got, err := store.Load()
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v err %v", got, err)
	}
}

func TestSavePermiso0600(t *testing.T) {
	dir := t.TempDir()
	store := New(dir)
	if err := store.Save(map[string]string{"K": "v"}); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(store.Path())
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("permiso %o, quiero 600", fi.Mode().Perm())
	}
}
