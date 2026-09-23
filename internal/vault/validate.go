package vault

import (
	"fmt"
	"strings"
)

func Validate(values map[string]string) error {
	for k := range values {
		if k == "" || strings.ContainsAny(k, "=\n\r#") || strings.TrimSpace(k) != k {
			return fmt.Errorf("nombre inválido: %q", k)
		}
	}
	return nil
}
