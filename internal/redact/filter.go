package redact

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

const nuke = "[REDACTED: posible secreto transformado]"

func Apply(data []byte, secrets []string) []byte {
	long := []string{}
	for _, s := range secrets {
		if len(s) >= 4 {
			long = append(long, s)
		}
	}
	if len(long) == 0 {
		return data
	}
	out := replaceLiterals(data, long)
	if scanEncoded(data, long) || scanSpaceless(out, long) {
		return []byte(nuke)
	}
	return out
}

func replaceLiterals(data []byte, secrets []string) []byte {
	patterns := []string{}
	for _, s := range secrets {
		patterns = append(patterns, s)
		for _, line := range strings.Split(s, "\n") {
			if line = strings.TrimSpace(line); len(line) >= 4 {
				patterns = append(patterns, line)
			}
		}
		hexed := hex.EncodeToString([]byte(s))
		uppered := strings.ToUpper(hexed)
		patterns = append(patterns,
			base64.StdEncoding.EncodeToString([]byte(s)),
			base64.RawStdEncoding.EncodeToString([]byte(s)),
			hexed,
			uppered,
			spacedHex(hexed),
			spacedHex(uppered),
			url.QueryEscape(s),
			reverse(s),
			strings.ToUpper(s),
			strings.ToLower(s),
			rot13(s),
		)
	}
	sort.Slice(patterns, func(i, j int) bool { return len(patterns[i]) > len(patterns[j]) })
	out := data
	for _, p := range patterns {
		if bytes.Contains(out, []byte(p)) {
			out = bytes.ReplaceAll(out, []byte(p), []byte("[REDACTED]"))
		}
	}
	return out
}

func scanEncoded(data []byte, secrets []string) bool {
	for _, token := range candidateTokens(string(data)) {
		for _, decoded := range decodeForms(token) {
			for _, s := range secrets {
				if len(s) >= 8 && strings.Contains(decoded, s) {
					return true
				}
			}
		}
	}
	return false
}

func candidateTokens(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		isToken := r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '+' || r == '/' || r == '='
		return !isToken
	})
	out := []string{}
	for _, f := range fields {
		if len(f) >= 16 && len(out) < 50 {
			chunk := f
			if len(chunk) > 4096 {
				chunk = chunk[:4096]
			}
			out = append(out, chunk)
		}
	}
	return out
}

func decodeForms(token string) []string {
	out := []string{}
	flat := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, token)
	if raw, err := base64.StdEncoding.DecodeString(flat); err == nil {
		out = append(out, string(raw))
	}
	if raw, err := base64.RawStdEncoding.DecodeString(flat); err == nil {
		out = append(out, string(raw))
	}
	if raw, err := hex.DecodeString(flat); err == nil {
		out = append(out, string(raw))
	}
	return out
}

func scanSpaceless(redacted []byte, secrets []string) bool {
	flat := stripSpaces(string(redacted))
	for _, s := range secrets {
		if strings.Contains(flat, stripSpaces(s)) {
			return true
		}
	}
	return false
}

func stripSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r', '-', '_', '.', ',', ':', ';', '/', '|':
			return -1
		}
		return r
	}, s)
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func rot13(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		default:
			return r
		}
	}, s)
}

func spacedHex(hexed string) string {
	pairs := make([]string, 0, len(hexed)/2)
	for i := 0; i+1 < len(hexed); i += 2 {
		pairs = append(pairs, hexed[i:i+2])
	}
	return strings.Join(pairs, " ")
}
