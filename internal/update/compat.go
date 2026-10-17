package update

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"regexp"
	"slices"
)

var pubKeyLiteral = regexp.MustCompile(`"([0-9a-f]{64})"`)

func KeysFromSource(src []byte) []string {
	var keys []string
	for _, m := range pubKeyLiteral.FindAllSubmatch(src, -1) {
		keys = append(keys, string(m[1]))
	}
	return keys
}

// CheckSigningCompat falla si la clave que firmará el próximo release no está
// entre las que ya conoce el release anterior: esos binarios rechazarían la
// firma y sus usuarios quedarían sin `hush update`.
func CheckSigningCompat(seedHex string, prevKeys []string) error {
	seed, err := hex.DecodeString(seedHex)
	if err != nil || len(seed) != ed25519.SeedSize {
		return fmt.Errorf("semilla inválida")
	}
	pub := hex.EncodeToString(ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey))
	if slices.Contains(prevKeys, pub) {
		return nil
	}
	return fmt.Errorf("la clave de firma %s… no está entre las %d que conoce el release anterior", pub[:8], len(prevKeys))
}
