package main

import (
	"os"

	"hush/internal/cli"
	"hush/internal/vault"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], vault.Default(), os.Stdin, os.Stdout, os.Stderr))
}
