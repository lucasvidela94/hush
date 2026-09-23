package vault

import "testing"

func TestValidate(t *testing.T) {
	bad := []map[string]string{
		{"": "x"},
		{"A=B": "x"},
		{"A\nB": "x"},
		{"A#B": "x"},
		{" A": "x"},
		{"A": "x\ny"},
		{"A": "x\ry"},
	}
	for _, values := range bad {
		if err := Validate(values); err == nil {
			t.Fatalf("aceptó %v", values)
		}
	}
	if err := Validate(map[string]string{"A_B-9.x": "ok"}); err != nil {
		t.Fatalf("rechazó válido: %v", err)
	}
}

func TestSaveRechazaCorrupto(t *testing.T) {
	store := New(t.TempDir())
	if err := store.Save(map[string]string{"K": "a\nb"}); err == nil {
		t.Fatal("guardó valor multilínea")
	}
}
