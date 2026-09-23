package redact

import (
	"bytes"
	"encoding/base32"
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
	if scanEncoded(string(data), long) || scanSpaceless(out, long) || scanStrided(out, long) || scanStridedRaw(string(data), long) || scanDecimal(string(data), long) {
		return []byte(nuke)
	}
	return out
}

func scanStrided(redacted []byte, secrets []string) bool {
	flat := []rune(normalize(string(redacted)))
	for _, s := range secrets {
		norm := normalize(s)
		if len(norm) < 4 {
			continue
		}
		for _, parity := range []int{0, 1} {
			var b strings.Builder
			for i := parity; i < len(flat); i += 2 {
				b.WriteRune(flat[i])
			}
			if strings.Contains(b.String(), norm) {
				return true
			}
		}
	}
	return false
}

func scanDecimal(s string, secrets []string) bool {
	fields := strings.Fields(s)
	var dec, oct []byte
	check := func(buf []byte) bool {
		if len(buf) < 4 {
			return false
		}
		for _, secret := range secrets {
			if len(secret) >= 8 && bytes.Contains(buf, []byte(secret)) {
				return true
			}
		}
		return false
	}
	for _, f := range fields {
		dn, on := 0, 0
		dok, ook := len(f) > 0 && len(f) <= 3, len(f) > 0 && len(f) <= 3
		for _, r := range f {
			if r < '0' || r > '9' {
				dok, ook = false, false
				break
			}
			dn = dn*10 + int(r-'0')
			if r > '7' {
				ook = false
			} else {
				on = on*8 + int(r-'0')
			}
		}
		if dok && dn <= 255 && len(dec) < 4096 {
			dec = append(dec, byte(dn))
		} else if check(dec) {
			return true
		} else {
			dec = nil
		}
		if ook && len(oct) < 4096 {
			oct = append(oct, byte(on))
		} else if check(oct) {
			return true
		} else {
			oct = nil
		}
	}
	return check(dec) || check(oct)
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

func scanEncoded(s string, secrets []string) bool {
	level := encodedRuns(s)
	for range 3 {
		next := []string{}
		for _, token := range level {
			for _, form := range decodeForms(token) {
				for _, secret := range secrets {
					if len(secret) >= 8 && (strings.Contains(form, secret) || strings.Contains(form, reverse(secret))) {
						return true
					}
				}
				if len(next) < 200 {
					next = append(next, form)
				}
			}
		}
		if len(next) == 0 {
			return false
		}
		level = next
	}
	return false
}

func encodedRuns(s string) []string {
	words := strings.Fields(s)
	runs := []string{}
	current := ""
	flush := func() {
		if len(current) >= 16 && len(runs) < 200 {
			if len(current) > 8192 {
				current = current[:8192]
			}
			runs = append(runs, current)
		}
		current = ""
	}
	for _, w := range words {
		subs := []string{}
		for _, sub := range strings.Split(w, "%") {
			subs = append(subs, splitEquals(sub)...)
		}
		isolated := len(subs) > 1
		if isolated {
			flush()
		}
		for _, part := range subs {
			if !isEncodedWord(part) {
				flush()
				continue
			}
			if len(part) >= 8 && len(part) <= 8192 && len(runs) < 200 && standsAlone(part) {
				runs = append(runs, part)
			}
			current += part
		}
		if isolated {
			flush()
		}
	}
	flush()
	return runs
}

func splitEquals(w string) []string {
	k := len(w)
	for k > 0 && w[k-1] == '=' {
		k--
	}
	head, tail := w[:k], w[k:]
	parts := strings.Split(head, "=")
	out := []string{}
	for i, p := range parts {
		if i == len(parts)-1 {
			p += tail
		}
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func isEncodedWord(w string) bool {
	if len(w) == 0 || len(w) > 256 {
		return false
	}
	for _, r := range w {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z',
			r == '+' || r == '/' || r == '-' || r == '_' || r == '=':
		default:
			return false
		}
	}
	return true
}

func standsAlone(w string) bool {
	letters := 0
	for _, r := range w {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			letters++
		}
	}
	return letters >= 4
}

func decodeForms(token string) []string {
	out := []string{}
	flat := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, token)
	b64 := strings.NewReplacer("-", "+", "_", "/").Replace(flat)
	for _, pad := range []string{"", "=", "=="} {
		if raw, err := base64.StdEncoding.DecodeString(b64 + pad); err == nil {
			out = append(out, string(raw))
		}
		if raw, err := base64.RawStdEncoding.DecodeString(b64); err == nil {
			out = append(out, string(raw))
		}
	}
	for _, pad := range []string{"", "=", "==", "===", "====", "====="} {
		if raw, err := base32.StdEncoding.DecodeString(strings.ToUpper(flat) + pad); err == nil {
			out = append(out, string(raw))
		}
	}
	if raw, err := hex.DecodeString(flat); err == nil {
		out = append(out, string(raw))
	}
	stripped := strings.Map(func(r rune) rune {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
			return r
		default:
			return -1
		}
	}, token)
	if len(stripped) >= 16 {
		if len(stripped)%2 == 1 {
			stripped = stripped[:len(stripped)-1]
		}
		if raw, err := hex.DecodeString(stripped); err == nil {
			out = append(out, string(raw))
		}
	}
	return out
}

func scanSpaceless(redacted []byte, secrets []string) bool {
	flat := normalize(string(redacted))
	for _, s := range secrets {
		norm := normalize(s)
		if len(norm) < 5 {
			continue
		}
		if strings.Contains(flat, norm) || strings.Contains(flat, reverse(norm)) {
			return true
		}
	}
	return false
}

func scanStridedRaw(data string, secrets []string) bool {
	text := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, data)
	flat := []rune(text)
	for _, s := range secrets {
		raw := strings.ToLower(s)
		if len(raw) < 8 {
			continue
		}
		for step := 2; step <= 4; step++ {
			for parity := 0; parity < step; parity++ {
				var b strings.Builder
				for i := parity; i < len(flat); i += step {
					b.WriteRune(flat[i])
				}
				if strings.Contains(b.String(), raw) {
					return true
				}
			}
		}
	}
	return false
}

func normalize(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		default:
			return -1
		}
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
