package redact

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

func Apply(data []byte, secrets []string) []byte {
	patterns := []string{}
	for _, s := range secrets {
		if len(s) < 4 {
			continue
		}
		patterns = append(patterns, s)
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
		)
	}
	sort.Slice(patterns, func(i, j int) bool { return len(patterns[i]) > len(patterns[j]) })
	out := data
	replaced := false
	for _, p := range patterns {
		if bytes.Contains(out, []byte(p)) {
			out = bytes.ReplaceAll(out, []byte(p), []byte("[REDACTED]"))
			replaced = true
		}
	}
	if replaced {
		return out
	}
	if hasSpacelessLeak(data, secrets) {
		return []byte("[REDACTED: posible secreto transformado]")
	}
	return out
}

func hasSpacelessLeak(data []byte, secrets []string) bool {
	flat := stripSpaces(string(data))
	for _, s := range secrets {
		if len(s) < 4 {
			continue
		}
		if strings.Contains(flat, stripSpaces(s)) {
			return true
		}
	}
	return false
}

func stripSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
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

func spacedHex(hexed string) string {
	pairs := make([]string, 0, len(hexed)/2)
	for i := 0; i+1 < len(hexed); i += 2 {
		pairs = append(pairs, hexed[i:i+2])
	}
	return strings.Join(pairs, " ")
}
