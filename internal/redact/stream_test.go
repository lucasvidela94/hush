package redact

import (
	"bytes"
	"strings"
	"testing"
)

func TestStreamNoFiltra(t *testing.T) {
	secret := "Zk9xQ7mPv2LwT8"
	long := strings.Repeat("Q7mPv2LwT8xZk9", 14)
	cases := []struct {
		name   string
		secret string
		writes []string
		limit  int
	}{
		{"token largo de a uno por línea", long, strings.SplitAfter(strings.Join(strings.Split(long[:40], ""), "\n")+"\n", "\n"), pendingLimit},
		{"literal en una línea", secret, []string{"key " + secret + "\n"}, pendingLimit},
		{"literal partido entre escrituras", secret, []string{"key Zk9xQ7", "mPv2LwT8\n"}, pendingLimit},
		{"un carácter por línea", secret, strings.SplitAfter(strings.Join(strings.Split(secret, ""), "\n")+"\n", "\n"), pendingLimit},
		{"base64 partido en líneas", secret, []string{"Wms5eFE3\n", "bVB2Mkx3\n", "VDg=\n"}, pendingLimit},
		{"sin salto y sin Flush previo", secret, []string{"key Zk9x", "Q7mPv2LwT8"}, pendingLimit},
		{"línea más larga que el límite", secret, []string{strings.Repeat("a", 60), "Zk9xQ7mP", "v2LwT8", strings.Repeat("b", 60) + "\n"}, 64},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			s := NewStream(&out, []string{tc.secret})
			s.limit = tc.limit
			for _, w := range tc.writes {
				if _, err := s.Write([]byte(w)); err != nil {
					t.Fatal(err)
				}
			}
			if err := s.Flush(); err != nil {
				t.Fatal(err)
			}
			flat := normalize(out.String())
			norm := normalize(tc.secret)
			n := min(windowLen(len(norm)), fragmentLen)
			for i := 0; i+n <= len(norm); i++ {
				if strings.Contains(flat, norm[i:i+n]) {
					t.Fatalf("quedó visible %q en %q", norm[i:i+n], out.String())
				}
			}
		})
	}
}

func TestStreamEntregaCadaLinea(t *testing.T) {
	var out bytes.Buffer
	s := NewStream(&out, []string{"Zk9xQ7mPv2LwT8"})
	if _, err := s.Write([]byte(`{"jsonrpc":"2.0","id":1}` + "\n" + `{"id":`)); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), `{"jsonrpc":"2.0","id":1}`+"\n"; got != want {
		t.Fatalf("antes de Flush got %q want %q", got, want)
	}
	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(out.String(), `{"id":`) {
		t.Fatalf("Flush no entregó la cola: %q", out.String())
	}
}

func TestStreamNoAlteraSalidaNormal(t *testing.T) {
	cases := []struct {
		name    string
		secrets []string
		data    string
	}{
		{"log común", []string{"Zk9xQ7mPv2LwT8"}, "build ok\n2 files changed\nlistening on port 8080\n"},
		{"sin secretos largos", []string{"abc"}, "abc\nsin salto"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			s := NewStream(&out, tc.secrets)
			for _, line := range strings.SplitAfter(tc.data, "\n") {
				if _, err := s.Write([]byte(line)); err != nil {
					t.Fatal(err)
				}
			}
			if err := s.Flush(); err != nil {
				t.Fatal(err)
			}
			if out.String() != tc.data {
				t.Fatalf("alteró %q → %q", tc.data, out.String())
			}
		})
	}
}
