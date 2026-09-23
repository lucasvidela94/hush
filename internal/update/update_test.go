package update

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v0.1.0", "v0.2.0", -1},
		{"v0.2.0", "v0.1.0", 1},
		{"v1.0.0", "v1.0.0", 0},
		{"v1.0.0", "v1.0.0-rc1", 1},
		{"0.1.22", "v0.1.22", 0},
	}
	for _, tc := range cases {
		if got := compareSemver(tc.a, tc.b); got != tc.want {
			t.Fatalf("compare(%q,%q)=%d quiero %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestNormalizeVersion(t *testing.T) {
	if got := normalizeVersion("0.1.2"); got != "v0.1.2" {
		t.Fatalf("got %q", got)
	}
	for _, bad := range []string{"", "dev", "v1.2", "hola"} {
		if got := normalizeVersion(bad); got != "" {
			t.Fatalf("%q → %q, quiero vacío", bad, got)
		}
	}
}

func TestChecksumFor(t *testing.T) {
	sums := "aaa  hush-linux-amd64.gz\nbbb  otro\n"
	got, err := checksumFor(sums, "hush-linux-amd64.gz")
	if err != nil || got != "aaa" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := checksumFor(sums, "ausente"); err == nil {
		t.Fatal("esperaba error")
	}
}

func TestGunzip(t *testing.T) {
	raw := []byte("hola")
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := gunzip(buf.Bytes())
	if err != nil || string(got) != "hola" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestDownloadAndVerifyEndToEnd(t *testing.T) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write([]byte("binario-falso")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(buf.Bytes())
	sums := hex.EncodeToString(sum[:]) + "  hush-linux-amd64.gz\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".gz"):
			_, _ = w.Write(buf.Bytes())
		case strings.HasSuffix(r.URL.Path, "checksums.txt"):
			_, _ = io.WriteString(w, sums)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	u := NewSelfUpdater()
	u.BaseURL = server.URL
	u.Repo = "cualquiera"
	raw, err := u.downloadAndVerify(context.Background(), "v9.9.9", "hush-linux-amd64.gz")
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "binario-falso" {
		t.Fatalf("got %q", raw)
	}
}

func TestSignVerifyRoundtrip(t *testing.T) {
	priv, pub, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("checksums de mentira")
	sig, err := SignBlob(priv, msg)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyBlob(pub, msg, sig); err != nil {
		t.Fatal(err)
	}
	if err := VerifyBlob(pub, []byte("otro"), sig); err == nil {
		t.Fatal("verificó mensaje alterado")
	}
}

func TestApplyReemplazaBinario(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "hush")
	if err := os.WriteFile(target, []byte("viejo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := apply(target, []byte("nuevo")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "nuevo" {
		t.Fatalf("got %q", got)
	}
}
