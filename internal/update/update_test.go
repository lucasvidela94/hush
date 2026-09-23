package update

import (
	"bytes"
	"compress/gzip"
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
