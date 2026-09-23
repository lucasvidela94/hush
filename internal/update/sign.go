package update

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateKey() (privHex, pubHex string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(priv.Seed()), hex.EncodeToString(pub), nil
}

func SignBlob(seedHex string, message []byte) (string, error) {
	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		return "", err
	}
	if len(seed) != ed25519.SeedSize {
		return "", fmt.Errorf("semilla inválida")
	}
	priv := ed25519.NewKeyFromSeed(seed)
	return hex.EncodeToString(ed25519.Sign(priv, message)), nil
}

func VerifyBlob(pubHex string, message []byte, sigHex string) error {
	pub, err := hex.DecodeString(pubHex)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("clave pública inválida")
	}
	sig, err := hex.DecodeString(sigHex)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("firma inválida")
	}
	if !ed25519.Verify(pub, message, sig) {
		return fmt.Errorf("la firma no coincide")
	}
	return nil
}
