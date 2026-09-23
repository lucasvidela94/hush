package vault

import (
	"fmt"
	"strings"
)

func Validate(values map[string]string) error {
	for k, v := range values {
		if k == "" || strings.ContainsAny(k, "=\n\r#") || strings.TrimSpace(k) != k {
			return fmt.Errorf("nombre inválido: %q", k)
		}
		if strings.ContainsAny(v, "\n\r") {
			return fmt.Errorf("el valor de %s tiene saltos de línea", k)
		}
	}
	return nil
}
