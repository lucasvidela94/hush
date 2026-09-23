package vault

import (
	"fmt"
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
	if isSymlink(s.Path()) {
		return nil, fmt.Errorf("el vault es un symlink, me niego a leerlo")
	}
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
		out[strings.TrimSpace(k)] = unescape(v)
	}
	return out, nil
}

func (s Store) Save(values map[string]string) error {
	if err := Validate(values); err != nil {
		return err
	}
	if isSymlink(s.Path()) {
		return fmt.Errorf("el vault es un symlink, me niego a escribirlo")
	}
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
		b.WriteString(k + "=" + escape(values[k]) + "\n")
	}
	tmp, err := os.CreateTemp(s.dir, "vault-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.WriteString(b.String()); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return err
	}
	if err := os.Chmod(s.dir, 0o700); err != nil {
		return err
	}
	return os.Rename(tmpName, s.Path())
}

func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

func escape(v string) string {
	v = strings.ReplaceAll(v, "\\", "\\\\")
	v = strings.ReplaceAll(v, "\n", "\\n")
	return strings.ReplaceAll(v, "\r", "\\r")
}

func unescape(v string) string {
	var b strings.Builder
	for i := 0; i < len(v); i++ {
		if v[i] == '\\' && i+1 < len(v) {
			switch v[i+1] {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case '\\':
				b.WriteByte('\\')
			default:
				b.WriteByte(v[i])
				b.WriteByte(v[i+1])
			}
			i++
			continue
		}
		b.WriteByte(v[i])
	}
	return b.String()
}
