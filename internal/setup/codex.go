package setup

import (
	"os"
	"path/filepath"
	"strings"
)

// codexRules va en un archivo propio: Codex carga todos los *.rules de
// ~/.codex/rules y gana la decisión más estricta, así que un allow que el
// usuario aprobó antes en default.rules no lo pisa.
const codexRules = `# Generado por hush setup: confirmá cada uso de secretos.
prefix_rule(
    pattern = ["hush", ["run", "stdin", "export"]],
    decision = "prompt",
    justification = "hush inyecta o muestra secretos",
)
`

// codexSandbox es lo que hush no escribe solo: config.toml es TOML y
// default_permissions va arriba de todo. Con el deny, el sandbox del SO
// bloquea ~/.hush también para hush corrido por shell; hush serve sigue
// andando porque Codex lanza los servidores MCP sin sandbox.
const codexSandbox = `default_permissions = "hush"   # arriba de todo, antes de cualquier [tabla]

[permissions.hush]
extends = ":workspace"

[permissions.hush.filesystem]
"~/.hush" = "deny"`

func codexRulesPath(home string) string {
	return filepath.Join(home, ".codex", "rules", "hush.rules")
}

func protectCodex(home string) (string, error) {
	path := codexRulesPath(home)
	state, err := writeCurrent(path, []byte(codexRules))
	if err != nil {
		return "", err
	}
	if state == "al día" {
		return "ya estaba protegido en " + path, nil
	}
	return "reglas en " + path + "\n  Para que el sandbox tampoco deje leer ~/.hush, agregá a ~/.codex/config.toml:\n" + indent(codexSandbox, "    "), nil
}

func protectedCodex(home string) bool {
	raw, err := os.ReadFile(codexRulesPath(home))
	return err == nil && string(raw) == codexRules
}

func codexSandboxCheck(home string) (Check, bool) {
	raw, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	if err != nil {
		return Check{}, false
	}
	for line := range strings.SplitSeq(string(raw), "\n") {
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if ok && (key == `"~/.hush"` || key == `"`+filepath.Join(home, ".hush")+`"`) && strings.Contains(value, `"deny"`) {
			return Check{Name: "sandbox codex", Status: OK}, true
		}
	}
	return Check{Name: "sandbox codex", Status: Warn, Detail: "el sandbox deja leer ~/.hush; agregá a ~/.codex/config.toml:\n" + codexSandbox}, true
}

func indent(s, prefix string) string {
	return prefix + strings.ReplaceAll(s, "\n", "\n"+prefix)
}
