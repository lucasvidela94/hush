package setup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"hush/internal/vault"
)

type Status int

const (
	OK Status = iota
	Warn
	Fail
)

type Check struct {
	Name   string
	Status Status
	Detail string
}

func Doctor(home string, store vault.Store, version, command string, stdout io.Writer) int {
	checks := runChecks(home, store, version, command)
	code := 0
	for _, c := range checks {
		mark := "✓"
		if c.Status == Warn {
			mark = "−"
		}
		if c.Status == Fail {
			mark = "✗"
			code = 1
		}
		if c.Detail == "" {
			fmt.Fprintf(stdout, "  %s %s\n", mark, c.Name)
		} else {
			fmt.Fprintf(stdout, "  %s %s: %s\n", mark, c.Name, c.Detail)
		}
	}
	return code
}

func runChecks(home string, store vault.Store, version, command string) []Check {
	checks := []Check{
		{Name: "binario", Status: OK, Detail: fmt.Sprintf("hush %s (%s)", version, command)},
	}
	checks = append(checks, vaultChecks(store)...)
	checks = append(checks, skillChecks(home)...)
	checks = append(checks, mcpChecks(home, command)...)
	return checks
}

func vaultChecks(store vault.Store) []Check {
	fi, err := os.Lstat(store.Path())
	if err != nil {
		if os.IsNotExist(err) {
			return []Check{{Name: "vault", Status: Warn, Detail: "vacío, cargá con hush set"}}
		}
		return []Check{{Name: "vault", Status: Fail, Detail: err.Error()}}
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return []Check{{Name: "vault", Status: Fail, Detail: "es un symlink"}}
	}
	if fi.Mode().Perm() != 0o600 {
		return []Check{{Name: "vault", Status: Fail, Detail: fmt.Sprintf("permiso %o, quiero 600", fi.Mode().Perm())}}
	}
	values, err := store.Load()
	if err != nil {
		return []Check{{Name: "vault", Status: Fail, Detail: err.Error()}}
	}
	return []Check{{Name: "vault", Status: OK, Detail: fmt.Sprintf("%s (%d secretos)", store.Path(), len(values))}}
}

func skillChecks(home string) []Check {
	checks := []Check{}
	for _, t := range Targets(home) {
		path := filepath.Join(t.Dir, "SKILL.md")
		raw, err := os.ReadFile(path)
		switch {
		case err != nil:
			checks = append(checks, Check{Name: "skill " + t.Harness, Status: Warn, Detail: "falta, corré hush setup"})
		case string(raw) != skill:
			checks = append(checks, Check{Name: "skill " + t.Harness, Status: Warn, Detail: "desactualizado, corré hush setup"})
		default:
			checks = append(checks, Check{Name: "skill " + t.Harness, Status: OK})
		}
	}
	return checks
}

func mcpChecks(home, command string) []Check {
	checks := []Check{}
	for _, h := range Harnesses() {
		switch h.Kind {
		case "json":
			checks = append(checks, jsonCheck(home, h))
		case "manual":
			checks = append(checks, Check{Name: "mcp " + h.Name, Status: Warn, Detail: "manual: " + h.Snippet(command)})
		default:
			checks = append(checks, Check{Name: "mcp " + h.Name, Status: Warn, Detail: h.Snippet(command)})
		}
	}
	return checks
}

func jsonCheck(home string, h Harness) Check {
	var path string
	switch h.Name {
	case "opencode":
		path = opencodePath(home)
	case "cursor":
		path = cursorPath(home)
	default:
		return Check{Name: "mcp " + h.Name, Status: Warn, Detail: "sin chequeo"}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Check{Name: "mcp " + h.Name, Status: Warn, Detail: "sin config, corré hush setup"}
	}
	if strings.Contains(string(raw), `"hush"`) {
		return Check{Name: "mcp " + h.Name, Status: OK, Detail: path}
	}
	return Check{Name: "mcp " + h.Name, Status: Warn, Detail: "sin entrada hush en " + path}
}
