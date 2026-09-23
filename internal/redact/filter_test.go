package redact

import (
	"strings"
	"testing"
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
		{"split con espacios", "token a b c 1 2 3", []string{"abc123"}, "[REDACTED: posible secreto transformado]"},
		{"mixto literal+split", "abc123 y a b c 1 2 3", []string{"abc123"}, "[REDACTED: posible secreto transformado]"},
		{"basic auth", "dXNlcjpzdXBlci1zZWNyZXQtd29yaw==", []string{"super-secret-work"}, "[REDACTED: posible secreto transformado]"},
		{"mayúsculas", "token ABC123", []string{"abc123"}, "token [REDACTED]"},
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

func TestApplyNoLeak(t *testing.T) {
	secret := "super-secreto-largo"
	got := string(Apply([]byte("echo "+secret), []string{secret}))
	if strings.Contains(got, secret) {
		t.Fatal("el valor quedó visible")
	}
}
