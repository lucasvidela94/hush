package main

import (
	"fmt"
	"os"

	"hush/internal/update"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: hush-sign gen | hush-sign sign <archivo>")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "gen":
		priv, pub, err := update.GenerateKey()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("HUSH_SIGN_PRIV=%s\nPUBKEY=%s\nGuardá la privada offline. Pegá la pública en internal/update/pubkey.go\n", priv, pub)
	case "sign":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "uso: hush-sign sign <archivo>")
			os.Exit(2)
		}
		seed := os.Getenv("HUSH_SIGN_PRIV")
		if seed == "" {
			fmt.Fprintln(os.Stderr, "hush-sign: falta HUSH_SIGN_PRIV")
			os.Exit(1)
		}
		raw, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		sig, err := update.SignBlob(seed, raw)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(sig)
	default:
		fmt.Fprintln(os.Stderr, "uso: hush-sign gen | hush-sign sign <archivo>")
		os.Exit(2)
	}
}
