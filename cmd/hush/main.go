package main

import (
	"os"

	"hush/internal/cli"
	"hush/internal/vault"
)

var version = "dev"

func main() {
	cli.Version = version
	os.Exit(cli.Run(os.Args[1:], vault.Default(), os.Stdin, os.Stdout, os.Stderr))
}
