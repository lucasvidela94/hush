package vault

import "testing"

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
