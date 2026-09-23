package setup

import (
	"os"
	"testing"
)

func TestApplyEscribeCuatroTargets(t *testing.T) {
	home := t.TempDir()
	results, err := Apply(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 4 {
		t.Fatalf("targets %d, quiero 4", len(results))
	}
	for _, r := range results {
		if r.State != "instalado" {
			t.Fatalf("%s: %s", r.Harness, r.State)
		}
		if _, err := os.Stat(r.Path); err != nil {
			t.Fatalf("%s: %v", r.Harness, err)
		}
	}
}

func TestApplyIdempotente(t *testing.T) {
	home := t.TempDir()
	if _, err := Apply(home); err != nil {
		t.Fatal(err)
	}
	results, err := Apply(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.State != "al día" {
			t.Fatalf("%s: %s", r.Harness, r.State)
		}
	}
}
