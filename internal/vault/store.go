package vault

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Store struct {
	dir string
}

func New(dir string) Store {
	return Store{dir: dir}
}

func Default() Store {
	home, _ := os.UserHomeDir()
	return New(filepath.Join(home, ".hush"))
}

func (s Store) Path() string {
	return filepath.Join(s.dir, "vault")
}

func (s Store) Load() (map[string]string, error) {
	out := map[string]string{}
	raw, err := os.ReadFile(s.Path())
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = v
	}
	return out, nil
}

func (s Store) Save(values map[string]string) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k + "=" + values[k] + "\n")
	}
	return os.WriteFile(s.Path(), []byte(b.String()), 0o600)
}
