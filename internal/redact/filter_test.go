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
