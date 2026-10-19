package redact

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"math"
	"net/url"
	"sort"
	"strings"
	"unicode"
)

const nuke = "[REDACTED: posible secreto transformado]"

const MinLength = 4

func Apply(data []byte, secrets []string) []byte {
	long := longSecrets(secrets)
	if len(long) == 0 {
		return data
	}
	return apply(data, long, literalPatterns(long))
}

func longSecrets(secrets []string) []string {
	long := []string{}
	for _, s := range secrets {
		if len(s) >= MinLength {
			long = append(long, s)
		}
	}
	return long
}

func apply(data []byte, secrets []string, patterns [][]byte) []byte {
	out := applyPatterns(data, patterns)
	if !transformed(data, out, secrets) {
		return out
	}
	return redactByLine(data, secrets, patterns)
}

func transformed(data, out []byte, secrets []string) bool {
	return scanEncoded(string(data), secrets) || scanSpaceless(out, secrets) || scanFragments(out, secrets) || scanStrided(out, secrets) || scanStridedRaw(string(data), secrets) || scanDecimal(string(data), secrets) || scanPermuted(out, secrets) || scanMasked(out, secrets)
}

func redactByLine(data []byte, secrets []string, patterns [][]byte) []byte {
	var kept bytes.Buffer
	flagged := false
	for _, line := range bytes.SplitAfter(data, []byte("\n")) {
		out := applyPatterns(line, patterns)
		if transformed(line, out, secrets) {
			flagged = true
			kept.WriteString(nuke + "\n")
			continue
		}
		kept.Write(out)
	}
	result := kept.Bytes()
	if !flagged || transformed(result, result, secrets) {
		return []byte(nuke + "\n")
	}
	return result
}

const (
	windowMin     = 6
	windowShare   = 0.4
	windowMaxLen  = 256
	scanMinLen    = 8
	maskedMatches = 0.6
	fragmentLen   = 12
)

func windowLen(n int) int {
	return max(windowMin, int(math.Ceil(float64(n)*windowShare)))
}

func windows(s string) []string {
	n := windowLen(len(s))
	if len(s) <= n || len(s) > windowMaxLen {
		return nil
	}
	out := make([]string, 0, len(s)-n+1)
	for i := 0; i+n <= len(s); i++ {
		out = append(out, s[i:i+n])
	}
	return out
}

func scanPermuted(redacted []byte, secrets []string) bool {
	flat := bytes.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, redacted)
	for _, s := range secrets {
		want := []byte(strings.Join(strings.Fields(s), ""))
		if len(want) < scanMinLen || len(flat) < len(want) {
			continue
		}
		var need, have [256]int
		for _, b := range want {
			need[b]++
		}
		diff := 0
		for _, n := range need {
			if n > 0 {
				diff++
			}
		}
		for i, b := range flat {
			switch have[b] {
			case need[b]:
				diff++
			case need[b] - 1:
				diff--
			}
			have[b]++
			if i >= len(want) {
				out := flat[i-len(want)]
				switch have[out] {
				case need[out]:
					diff++
				case need[out] + 1:
					diff--
				}
				have[out]--
			}
			if i >= len(want)-1 && diff == 0 {
				return true
			}
		}
	}
	return false
}

func scanMasked(redacted []byte, secrets []string) bool {
	for _, s := range secrets {
		if len(s) < scanMinLen || len(s) > windowMaxLen {
			continue
		}
		budget := len(s) - int(math.Ceil(float64(len(s))*maskedMatches))
		for i := 0; i+len(s) <= len(redacted); i++ {
			miss := 0
			for j := 0; j < len(s) && miss <= budget; j++ {
				if redacted[i+j] != s[j] {
					miss++
				}
			}
			if miss <= budget {
				return true
			}
		}
	}
	return false
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
		switch {
		case dok && dn <= 255 && len(dec) < 4096:
			dec = append(dec, byte(dn))
		case check(dec):
			return true
		default:
			dec = nil
		}
		switch {
		case ook && len(oct) < 4096:
			oct = append(oct, byte(on))
		case check(oct):
			return true
		default:
			oct = nil
		}
	}
	return check(dec) || check(oct)
}

func applyPatterns(data []byte, patterns [][]byte) []byte {
	out := data
	for _, p := range patterns {
		if bytes.Contains(out, p) {
			out = bytes.ReplaceAll(out, p, []byte("[REDACTED]"))
		}
	}
	return out
}

func literalPatterns(secrets []string) [][]byte {
	patterns := []string{}
	for _, s := range secrets {
		patterns = append(patterns, s)
		patterns = append(patterns, windows(s)...)
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
	raw := make([][]byte, len(patterns))
	for i, p := range patterns {
		raw[i] = []byte(p)
	}
	return raw
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
	const maxRun = 8192
	var current strings.Builder
	flush := func() {
		if current.Len() >= 16 && len(runs) < 200 {
			runs = append(runs, current.String())
		}
		current.Reset()
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
			if room := maxRun - current.Len(); room > 0 {
				current.WriteString(part[:min(room, len(part))])
			}
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

// scanFragments busca un tramo corto del secreto con los separadores quitados:
// detecta uno que sale de a pedazos o recortado antes de que esté completo.
// El tramo es fijo para que de un token largo no se escape casi la mitad.
func scanFragments(redacted []byte, secrets []string) bool {
	flat := normalize(strings.NewReplacer(nuke, " ", "[REDACTED]", " ").Replace(string(redacted)))
	for _, s := range secrets {
		norm := normalize(s)
		n := min(windowLen(len(norm)), fragmentLen)
		if len(norm) <= n || len(flat) < n {
			continue
		}
		frags := make(map[string]bool, 2*len(norm))
		for i := 0; i+n <= len(norm); i++ {
			frags[norm[i:i+n]] = true
			frags[reverse(norm[i:i+n])] = true
		}
		for i := 0; i+n <= len(flat); i++ {
			if frags[flat[i:i+n]] {
				return true
			}
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
