package setup

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func opencodePath(home string) string {
	return filepath.Join(home, ".config", "opencode", "opencode.json")
}

// opencodeVersion dice qué formato usa la config: 1, 2, o 0 si no se puede
// saber. Las marcas de la config mandan; sin marcas, la versión instalada.
func opencodeVersion(config map[string]any) int {
	if _, ok := config["permission"]; ok {
		return 1
	}
	if _, ok := config["permissions"]; ok {
		return 2
	}
	if mcp, ok := config["mcp"].(map[string]any); ok {
		if _, ok := mcp["servers"]; ok {
			return 2
		}
	}
	return installedOpencodeMajor()
}

func installedOpencodeMajor() int {
	path, err := exec.LookPath("opencode")
	if err != nil {
		return 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "--version").Output()
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return 0
	}
	major, _, _ := strings.Cut(strings.TrimPrefix(fields[len(fields)-1], "v"), ".")
	switch major {
	case "1":
		return 1
	case "2":
		return 2
	}
	return 0
}

// applyOpencode registra el servidor: v2 en mcp.servers.hush (y mueve una
// entrada vieja de mcp.hush), v1 o desconocida en mcp.hush.
func applyOpencode(home string, command string) (string, error) {
	path := opencodePath(home)
	changed, err := editJSON(path, func(config map[string]any) (bool, error) {
		entry := map[string]any{"type": "local", "command": []string{command, "serve"}}
		mcp, ok := config["mcp"].(map[string]any)
		if _, exists := config["mcp"]; exists && !ok {
			return false, errors.New("mcp no es un objeto, no lo toco")
		}
		if mcp == nil {
			mcp = map[string]any{}
		}
		config["mcp"] = mcp
		if opencodeVersion(config) != 2 {
			if _, ok := mcp["hush"]; ok {
				return false, nil
			}
			mcp["hush"] = entry
			return true, nil
		}
		servers, ok := mcp["servers"].(map[string]any)
		if _, exists := mcp["servers"]; exists && !ok {
			return false, errors.New("mcp.servers no es un objeto, no lo toco")
		}
		if servers == nil {
			servers = map[string]any{}
		}
		mcp["servers"] = servers
		_, stale := mcp["hush"]
		delete(mcp, "hush")
		if _, ok := servers["hush"]; ok {
			return stale, nil
		}
		servers["hush"] = entry
		return true, nil
	})
	if err != nil {
		return "", err
	}
	if !changed {
		return "ya estaba registrado", nil
	}
	return "registrado en " + path, nil
}

func opencodeCheck(home string) Check {
	path := opencodePath(home)
	config, raw, err := loadJSON(path)
	if err != nil {
		return Check{Name: "mcp opencode", Status: Warn, Detail: err.Error()}
	}
	if raw == nil {
		return Check{Name: "mcp opencode", Status: Warn, Detail: "sin config, corré hush setup"}
	}
	mcp, _ := config["mcp"].(map[string]any)
	_, legacy := mcp["hush"]
	if opencodeVersion(config) != 2 {
		if legacy {
			return Check{Name: "mcp opencode", Status: OK, Detail: path}
		}
		return Check{Name: "mcp opencode", Status: Warn, Detail: "sin entrada hush en " + path}
	}
	servers, _ := mcp["servers"].(map[string]any)
	if _, ok := servers["hush"]; ok {
		return Check{Name: "mcp opencode", Status: OK, Detail: path}
	}
	if legacy {
		return Check{Name: "mcp opencode", Status: Warn, Detail: "entrada hush en formato v1 (mcp.hush), corré hush setup"}
	}
	return Check{Name: "mcp opencode", Status: Warn, Detail: "sin entrada hush en " + path}
}
