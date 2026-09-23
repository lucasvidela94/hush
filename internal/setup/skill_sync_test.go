package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillSincronizado(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != skill {
		t.Fatal("internal/setup/skill.md desactualizado: corré go generate ./...")
	}
}
