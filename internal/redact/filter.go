package redact

import "bytes"

func Apply(data []byte, secrets []string) []byte {
	out := data
	for _, s := range secrets {
		if len(s) >= 4 {
			out = bytes.ReplaceAll(out, []byte(s), []byte("[REDACTED]"))
		}
	}
	return out
}
