package setup

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
)

// Las reglas de protección le impiden al agente leer o editar ~/.hush con
// sus herramientas y le piden confirmación humana antes de inyectar un
// secreto. Son fricción contra un agente con shell, no un límite: ver
// SECURITY.md.

func claudeSettingsPath(home string) string {
	return filepath.Join(home, ".claude", "settings.json")
}

func claudeRules(kind string) []string {
	switch kind {
	case "deny":
		return []string{"Read(~/.hush/**)", "Edit(~/.hush/**)"}
	case "ask":
		return []string{"Bash(hush run *)", "Bash(hush stdin *)", "Bash(hush export *)"}
	}
	return nil
}

func mergeClaude(config map[string]any) (bool, error) {
	perms := map[string]any{}
	if raw, ok := config["permissions"]; ok {
		if perms, ok = raw.(map[string]any); !ok {
			return false, errors.New("permissions no es un objeto, no lo toco")
		}
	}
	changed := false
	for _, kind := range []string{"deny", "ask"} {
		list, ok := perms[kind].([]any)
		if _, exists := perms[kind]; exists && !ok {
			return false, fmt.Errorf("permissions.%s no es una lista, no lo toco", kind)
		}
		for _, rule := range claudeRules(kind) {
			if !slices.Contains(list, any(rule)) {
				list = append(list, rule)
				changed = true
			}
		}
		perms[kind] = list
	}
	config["permissions"] = perms
	return changed, nil
}

// opencodeRules va al final de permissions: en opencode v2 gana la última
// regla que coincide. El shell no expande ~, por eso el patrón literal.
func opencodeRules() []map[string]any {
	rule := func(action, resource, effect string) map[string]any {
		return map[string]any{"action": action, "resource": resource, "effect": effect}
	}
	return []map[string]any{
		rule("external_directory", "~/.hush/*", "deny"),
		rule("read", "~/.hush/*", "deny"),
		rule("edit", "~/.hush/*", "deny"),
		rule("shell", "*.hush/*", "deny"),
		rule("shell", "hush run *", "ask"),
		rule("shell", "hush stdin *", "ask"),
		rule("shell", "hush export *", "ask"),
	}
}

func mergeOpencode(config map[string]any) (bool, error) {
	if _, v1 := config["permission"]; v1 {
		return false, errors.New("config de opencode v1 (permission), agregá las reglas a mano: ver SECURITY.md")
	}
	rules, isList := config["permissions"].([]any)
	if _, exists := config["permissions"]; exists && !isList {
		return false, errors.New("permissions no es una lista, no lo toco")
	}
	mcp, _ := config["mcp"].(map[string]any)
	if _, servers := mcp["servers"]; !isList && !servers {
		return false, errors.New("no sé si esta config es de opencode v1 o v2, agregá las reglas a mano: ver SECURITY.md")
	}
	changed := false
	for _, want := range opencodeRules() {
		if !slices.ContainsFunc(rules, func(r any) bool { return sameRule(r, want) }) {
			rules = append(rules, want)
			changed = true
		}
	}
	config["permissions"] = rules
	return changed, nil
}

func sameRule(r any, want map[string]any) bool {
	m, ok := r.(map[string]any)
	return ok && m["action"] == want["action"] && m["resource"] == want["resource"] && m["effect"] == want["effect"]
}

func protectWith(pathOf func(string) string, merge func(map[string]any) (bool, error)) func(string) (string, error) {
	return func(home string) (string, error) {
		path := pathOf(home)
		changed, err := editJSON(path, merge)
		if err != nil {
			return "", err
		}
		if !changed {
			return "ya estaba protegido en " + path, nil
		}
		return "reglas agregadas en " + path, nil
	}
}

func protectedWith(pathOf func(string) string, merge func(map[string]any) (bool, error)) func(string) bool {
	return func(home string) bool {
		config, _, err := loadJSON(pathOf(home))
		if err != nil {
			return false
		}
		changed, err := merge(config)
		return err == nil && !changed
	}
}
