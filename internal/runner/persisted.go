package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"hush/internal/redact"
)

const persistedReadLimit = 4 << 20

type stamp struct {
	mod  time.Time
	size int64
}

type persisted struct {
	name string
	path string
}

func watchedPaths(home, cwd string) []string {
	paths := []string{
		filepath.Join(cwd, ".mcp.json"),
		filepath.Join(cwd, "opencode.json"),
	}
	if home != "" {
		paths = append(paths,
			filepath.Join(home, ".claude.json"),
			filepath.Join(home, ".cursor", "mcp.json"),
			filepath.Join(home, ".config", "opencode", "opencode.json"),
		)
	}
	envFiles, _ := filepath.Glob(filepath.Join(cwd, ".env*"))
	return append(paths, envFiles...)
}

func snapshot(paths []string) map[string]stamp {
	out := map[string]stamp{}
	for _, p := range paths {
		if fi, err := os.Stat(p); err == nil && fi.Mode().IsRegular() {
			out[p] = stamp{mod: fi.ModTime(), size: fi.Size()}
		}
	}
	return out
}

func findPersisted(before, after map[string]stamp, secrets map[string]string) []persisted {
	var found []persisted
	for path, now := range after {
		if was, ok := before[path]; ok && was == now {
			continue
		}
		content := readCapped(path)
		for name, value := range secrets {
			if len(value) >= redact.MinLength && bytes.Contains(content, []byte(value)) {
				found = append(found, persisted{name: name, path: path})
			}
		}
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].path != found[j].path {
			return found[i].path < found[j].path
		}
		return found[i].name < found[j].name
	})
	return found
}

func readCapped(path string) []byte {
	f, err := os.Open(path) //nolint:gosec // ruta de la lista fija de archivos vigilados
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()
	data, _ := io.ReadAll(io.LimitReader(f, persistedReadLimit))
	return data
}

func warnPersisted(stderr io.Writer, found []persisted) {
	for _, f := range found {
		fmt.Fprintf(stderr, "hush: aviso: %s quedó guardado en texto plano en %s; rotalo o usá una referencia a variable de entorno, y no lo leas de vuelta fuera de hush\n", f.name, f.path)
	}
}
