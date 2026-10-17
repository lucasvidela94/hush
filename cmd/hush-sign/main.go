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
	case "compat":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "uso: hush-sign compat <pubkey.go del release anterior>")
			os.Exit(2)
		}
		src, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := update.CheckSigningCompat(os.Getenv("HUSH_SIGN_PRIV"), update.KeysFromSource(src)); err != nil {
			fmt.Fprintf(os.Stderr, "hush-sign: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("clave de firma compatible con el release anterior")
	default:
		fmt.Fprintln(os.Stderr, "uso: hush-sign gen | hush-sign sign <archivo> | hush-sign compat <pubkey.go>")
		os.Exit(2)
	}
}
