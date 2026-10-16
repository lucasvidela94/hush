package redact

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestApply(t *testing.T) {
	cases := []struct {
		name    string
		data    string
		secrets []string
		want    string
	}{
		{"reemplaza valor", "la key es abc123 fin", []string{"abc123"}, "la key es [REDACTED] fin"},
		{"ignora cortos", "abc xyz", []string{"abc"}, "abc xyz"},
		{"sin secretos", "hola", nil, "hola"},
		{"multiples", "a1b2 x3y4", []string{"a1b2", "x3y4"}, "[REDACTED] [REDACTED]"},
		{"base64", "token YWJjMTIz", []string{"abc123"}, "token [REDACTED]"},
		{"reversa", "token 321cba", []string{"abc123"}, "token [REDACTED]"},
		{"hex", "token 616263313233", []string{"abc123"}, "token [REDACTED]"},
		{"hex espaciado", "token 61 62 63 31 32 33", []string{"abc123"}, "token [REDACTED]"},
		{"split con espacios", "token a b c 1 2 3", []string{"abc123"}, "[REDACTED: posible secreto transformado]\n"},
		{"mixto literal+split", "abc123 y a b c 1 2 3", []string{"abc123"}, "[REDACTED: posible secreto transformado]\n"},
		{"basic auth", "dXNlcjpzdXBlci1zZWNyZXQtd29yaw==", []string{"super-secret-work"}, "[REDACTED: posible secreto transformado]\n"},
		{"mayúsculas", "token ABC123", []string{"abc123"}, "token [REDACTED]"},
		{"mayúsculas + espaciado", "token A B C D E F 1 2", []string{"abcdef12"}, "[REDACTED: posible secreto transformado]\n"},
		{"sin alfanuméricos no anula", "hola mundo", []string{"!@#$%^&*()"}, "hola mundo"},
		{"corto con espacios pasa", "token a b c", []string{"abc"}, "token a b c"},
		{"base32", "token ONSWG4TFOQYTEMY=", []string{"secret123"}, "[REDACTED: posible secreto transformado]\n"},
		{"decimal por byte", "token 115 101 99 114 101 116 49 50 51", []string{"secret123"}, "[REDACTED: posible secreto transformado]\n"},
		{"hex 2char largo", "token 6d 79 73 65 63 72 65 74 6b 65 79", []string{"mysecretkey"}, "token [REDACTED]"},
		{"intercalado", "token sZ3ZcZrZ3ZtZ9Z9Z", []string{"s3cr3t99"}, "[REDACTED: posible secreto transformado]\n"},
		{"reverso espaciado", "token 3 2 1 T E R C E S R E P U S", []string{"SUPERSECRET123"}, "[REDACTED: posible secreto transformado]\n"},
		{"sk_live stride", "token sZkZ_ZlZiZvZeZ_ZSZUZPZEZRZSZEZCZRZEZTZ1Z2Z3Z", []string{"sk_live_SUPERSECRET123"}, "[REDACTED: posible secreto transformado]\n"},
		{"percent-hex", "token %6d%79%73%65%63%72%65%74%6b%65%79", []string{"mysecretkey"}, "[REDACTED: posible secreto transformado]\n"},
		{"octal", "token 155 171 163 145 143 162 145 164 153 145 171", []string{"mysecretkey"}, "[REDACTED: posible secreto transformado]\n"},
		{"base64 del reverso", "token MzIxdGVyY2Vz", []string{"secret123"}, "[REDACTED: posible secreto transformado]\n"},
		{"hex del reverso", "token 3332312d646c726f772d6f6c6c6568", []string{"hello-world-123"}, "[REDACTED: posible secreto transformado]\n"},
		{"token= con prefijo", "token=dXNlcjpzdXBlci1zZWNyZXQtd29yaw==", []string{"super-secret-work"}, "[REDACTED: posible secreto transformado]\n"},
		{"base64url", "token=Pj4-Pz8_Pj4-Pz8_", []string{">>>???>>>???"}, "[REDACTED: posible secreto transformado]\n"},
		{"base64 con wrap", "dXNlcjpzdXBl\nci1zZWNyZXQtd29yaw==", []string{"super-secret-work"}, "[REDACTED: posible secreto transformado]\n"},
		{"doble base64", "ZFhObGNqcHpkWEJsY2kxelpXTnlaWFF0ZDI5eWF3PT0=", []string{"super-secret-work"}, "[REDACTED: posible secreto transformado]\n"},
		{"substring ordenado", "largo y larg", []string{"larg", "largo"}, "[REDACTED] y [REDACTED]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(Apply([]byte(tc.data), tc.secrets))
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestApplyPorLinea(t *testing.T) {
	const marker = "[REDACTED: posible secreto transformado]\n"
	secret := "super-secret-work"
	cases := []struct {
		name string
		data string
		want string
	}{
		{"conserva lo no secreto", "A\ndXNlcjpzdXBlci1zZWNyZXQtd29yaw==\nB\n", "A\n" + marker + "B\n"},
		{"sin newline final", "A\ndXNlcjpzdXBlci1zZWNyZXQtd29yaw==", "A\n" + marker},
		{"literal y transformado", "uno " + secret + "\ndXNlcjpzdXBlci1zZWNyZXQtd29yaw==\n", "uno [REDACTED]\n" + marker},
		{"repartido en líneas cae a todo", "dXNlcjpzdXBl\nci1zZWNyZXQtd29yaw==\n", marker},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(Apply([]byte(tc.data), []string{secret}))
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestApplySubstringPermutacionYMascara(t *testing.T) {
	const marker = "[REDACTED: posible secreto transformado]\n"
	secret := "Zk9xQ7mPv2LwT8"
	cases := []struct {
		name   string
		data   string
		hidden []string
	}{
		{"cut -c3-9", "9xQ7mPv\n", []string{"9xQ7mP", "xQ7mPv"}},
		{"ventana de 6", "ver: Zk9xQ7 fin\n", []string{"Zk9xQ7"}},
		{"permutación ordenada", "2789LPQTZkmvwx\n", []string{"2789LPQTZkmvwx"}},
		{"permutación uno por línea", "2\n7\n8\n9\nL\nP\nQ\nT\nZ\nk\nm\nv\nw\nx\n", nil},
		{"máscara de dígitos", "Zk#xQ#mPv#LwT#\n", []string{"Zk#xQ#mPv#LwT#"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(Apply([]byte(tc.data), []string{secret}))
			if got == tc.data {
				t.Fatalf("pasó sin tocar: %q", got)
			}
			for _, h := range tc.hidden {
				if strings.Contains(got, h) {
					t.Fatalf("quedó visible %q en %q", h, got)
				}
			}
			if strings.Contains(got, "REDACTED") == false && got != marker {
				t.Fatalf("sin marcador: %q", got)
			}
		})
	}
}

func TestApplyNoFalsosPositivos(t *testing.T) {
	secret := "Zk9xQ7mPv2LwT8"
	for _, data := range []string{
		"build ok\n2 files changed\n",
		"Zk9x\n",
		"listening on port 8080\n",
		"qwertyuiopasdf\n",
	} {
		if got := string(Apply([]byte(data), []string{secret})); got != data {
			t.Errorf("alteró %q → %q", data, got)
		}
	}
}

func TestApplySalidaGrandeEsRapida(t *testing.T) {
	secrets := []string{"Zk9xQ7mPv2LwT8", "sk_live_SUPERSECRET123456789"}
	normal := strings.Repeat("linea de log normal sin nada raro 12345\n", 1<<20/40)
	cases := map[string]string{
		"normal":                  normal,
		"con una línea de base64": normal + "c2tfbGl2ZV9TVVBFUlNFQ1JFVDEyMzQ1Njc4OQ==\n",
		"sin saltos de línea":     strings.ReplaceAll(normal, "\n", " "),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			got := Apply([]byte(data), secrets)
			if d := time.Since(start); d > 10*time.Second {
				t.Fatalf("tardó %s", d)
			}
			if bytes.Contains(got, []byte("sk_live_SUPERSECRET")) {
				t.Fatal("quedó visible")
			}
		})
	}
}

func BenchmarkApplyUnMegabyte(b *testing.B) {
	data := []byte(strings.Repeat("linea de log normal sin nada raro 12345\n", 1<<20/40) + "c2tfbGl2ZV9TVVBFUlNFQ1JFVDEyMzQ1Njc4OQ==\n")
	secrets := []string{"Zk9xQ7mPv2LwT8", "sk_live_SUPERSECRET123456789"}
	for b.Loop() {
		Apply(data, secrets)
	}
}

func TestApplyNoLeak(t *testing.T) {
	secret := "super-secreto-largo"
	got := string(Apply([]byte("echo "+secret), []string{secret}))
	if strings.Contains(got, secret) {
		t.Fatal("el valor quedó visible")
	}
}
