package update

import (
	"os"
	"testing"
)

func TestKeysFromSourceLeeLaListaReal(t *testing.T) {
	src, err := os.ReadFile("pubkey.go")
	if err != nil {
		t.Fatal(err)
	}
	got := KeysFromSource(src)
	if len(got) != len(releasePubKeys) {
		t.Fatalf("leyó %d claves, hay %d", len(got), len(releasePubKeys))
	}
	for i := range got {
		if got[i] != releasePubKeys[i] {
			t.Fatalf("clave %d distinta", i)
		}
	}
}

func TestCheckSigningCompat(t *testing.T) {
	privA, pubA, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	privB, pubB, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		priv string
		prev []string
		ok   bool
	}{
		{"misma clave", privA, []string{pubA}, true},
		{"clave nueva conocida por el anterior", privB, []string{pubA, pubB}, true},
		{"rotación sin puente", privB, []string{pubA}, false},
		{"anterior sin claves", privA, nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := CheckSigningCompat(c.priv, c.prev)
			if (err == nil) != c.ok {
				t.Fatalf("err=%v, ok esperado %v", err, c.ok)
			}
		})
	}
	if err := CheckSigningCompat("zz", []string{pubA}); err == nil {
		t.Fatal("aceptó una semilla inválida")
	}
}
