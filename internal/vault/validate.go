package vault

import (
	"fmt"
	"regexp"
	"strings"
)

var envName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func ValidName(name string) error {
	if !envName.MatchString(name) {
		return fmt.Errorf("nombre inválido %q: usá letras, dígitos y _, sin empezar con dígito", name)
	}
	return nil
}

func Validate(values map[string]string) error {
	for k := range values {
		if k == "" || strings.ContainsAny(k, "=\n\r#") || strings.TrimSpace(k) != k {
			return fmt.Errorf("nombre inválido: %q", k)
		}
	}
	return nil
}
