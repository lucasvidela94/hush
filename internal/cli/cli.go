package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"hush/internal/mcp"
	"hush/internal/runner"
	"hush/internal/secret"
	"hush/internal/setup"
	"hush/internal/update"
	"hush/internal/vault"
)

const (
	OK      = 0
	Failed  = 1
	Usage   = 2
	Missing = 3
)

var Version = "dev"

func Run(argv []string, store vault.Store, stdin *os.File, stdout, stderr io.Writer) int {
	if len(argv) == 0 {
		printUsage(stdout)
		return Usage
	}
	switch argv[0] {
	case "set":
		return set(argv[1:], store, stdin, stdout)
	case "check":
		return check(argv[1:], store, stdout, stderr)
	case "list":
		return list(store, stdout, stderr)
	case "run":
		return run(argv[1:], store, stdin, stdout, stderr)
	case "stdin":
		return pipeIn(argv[1:], store, stdout, stderr)
	case "export":
		return exportSecrets(argv[1:], store, stdout, stderr)
	case "serve":
		return serve(store, stderr)
	case "setup":
		return runSetup(stdout, stderr)
	case "update":
		return runUpdate(stdout, stderr)
	case "status":
		return status(store, stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "hush %s\n", Version)
		return OK
	case "help", "--help", "-h":
		return help(argv[1:], stdout)
	default:
		fmt.Fprintf(stderr, "hush: comando desconocido %q\n\n", argv[0])
		printUsage(stdout)
		return Usage
	}
}

func set(argv []string, store vault.Store, stdin *os.File, stdout io.Writer) int {
	if len(argv) != 1 {
		fmt.Fprintln(stdout, "uso: hush set NOMBRE")
		return Usage
	}
	value, err := secret.Read(stdin)
	if err != nil || value == "" {
		fmt.Fprintln(stdout, "hush: valor vacío, cancelado")
		return Failed
	}
	values, err := store.Load()
	if err != nil {
		fmt.Fprintln(stdout, "hush: no se pudo leer el vault")
		return Failed
	}
	values[argv[0]] = value
	if err := store.Save(values); err != nil {
		fmt.Fprintf(stdout, "hush: no se pudo guardar: %s\n", err)
		return Failed
	}
	fmt.Fprintf(stdout, "hush: %s guardado (0600)\n", argv[0])
	return OK
}

func check(argv []string, store vault.Store, stdout, stderr io.Writer) int {
	names := []string{}
	asJSON := false
	for _, a := range argv {
		if a == "--json" {
			asJSON = true
			continue
		}
		names = append(names, a)
	}
	if len(names) == 0 {
		fmt.Fprintln(stderr, "uso: hush check [--json] NOMBRE...")
		return Usage
	}
	values, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, "hush: no se pudo leer el vault")
		return Failed
	}
	missing := []string{}
	for _, n := range names {
		if _, ok := values[n]; !ok {
			missing = append(missing, n)
		}
	}
	if asJSON {
		raw, _ := json.Marshal(map[string][]string{"missing": missing})
		fmt.Fprintln(stdout, string(raw))
		if len(missing) > 0 {
			return Missing
		}
		return OK
	}
	if len(missing) == 0 {
		fmt.Fprintln(stdout, "hush: ok, todo presente")
		return OK
	}
	fmt.Fprintf(stderr, "hush: faltan: %s → pedí al humano: hush set %s\n", strings.Join(missing, ", "), missing[0])
	return Missing
}

func list(store vault.Store, stdout, stderr io.Writer) int {
	values, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, "hush: no se pudo leer el vault")
		return Failed
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	for _, k := range keys {
		fmt.Fprintln(stdout, k)
	}
	return OK
}

func run(argv []string, store vault.Store, stdin *os.File, stdout, stderr io.Writer) int {
	only := map[string]bool{}
	rest := argv
	if len(argv) >= 1 && argv[0] == "--only" {
		if len(argv) < 3 {
			fmt.Fprintln(stderr, "uso: hush run [--only A,B] -- comando...")
			return Usage
		}
		for _, k := range strings.Split(argv[1], ",") {
			only[strings.TrimSpace(k)] = true
		}
		rest = argv[2:]
	}
	if len(rest) > 0 && rest[0] == "--" {
		rest = rest[1:]
	}
	if len(rest) == 0 {
		fmt.Fprintln(stderr, "uso: hush run [--only A,B] -- comando...")
		return Usage
	}
	values, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, "hush: no se pudo leer el vault")
		return Failed
	}
	extra := map[string]string{}
	for k, v := range values {
		if len(only) == 0 || only[k] {
			extra[k] = v
		}
	}
	return runner.Run(extra, rest, stdin, stdout, stderr)
}

func pipeIn(argv []string, store vault.Store, stdout, stderr io.Writer) int {
	if len(argv) < 3 || argv[1] != "--" {
		fmt.Fprintln(stderr, "uso: hush stdin NOMBRE -- comando...")
		return Usage
	}
	values, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, "hush: no se pudo leer el vault")
		return Failed
	}
	value, ok := values[argv[0]]
	if !ok {
		fmt.Fprintf(stderr, "hush: falta %s → pedí al humano: hush set %s\n", argv[0], argv[0])
		return Missing
	}
	return runner.Pipe(value, argv[2:], stdout, stderr)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `hush — tus secretos, sin que el agente los vea.

  hush set NOMBRE                  guarda un valor (te lo pide sin mostrarlo)
  hush check NOMBRE...             ¿están guardados? (solo nombres)
  hush list                        qué nombres hay guardados
  hush run [--only A,B] -- cmd...  corre un comando con los valores inyectados
  hush stdin NOMBRE -- cmd...      le escribe un valor al stdin del comando
  hush export [NOMBRES...]         muestra valores SOLO en terminal real
  hush serve                       servidor MCP para tu harness de IA
  hush setup                       instala el skill en tus harnesses
  hush update                      actualiza hush
  hush status                      versión + vault + skills
  hush version                     versión
  hush help [COMANDO]              ayuda de un comando

ej:
  openssl rand -hex 24 | hush set WHATSAPP_VERIFY_TOKEN
  hush stdin WHATSAPP_VERIFY_TOKEN -- npx wrangler secret put WHATSAPP_VERIFY_TOKEN`)
}

var helpText = map[string]string{
	"set":    "uso: hush set NOMBRE\nej: openssl rand -hex 24 | hush set MI_TOKEN\nGuarda un valor leyéndolo de tu terminal (sin mostrarlo) o de un pipe.",
	"check":  "uso: hush check [--json] NOMBRE...\nDice qué nombres están guardados y cuáles faltan. Nunca muestra valores.",
	"list":   "uso: hush list\nLista los nombres guardados. Nunca muestra valores.",
	"run":    "uso: hush run [--only A,B] -- comando...\nCorre el comando con los secretos como variables de entorno y tapa los valores en la salida.\nej: hush run --only API_KEY -- ./deploy.sh",
	"stdin":  "uso: hush stdin NOMBRE -- comando...\nLe escribe el valor al stdin del comando. Para programas que piden el secreto por consola.\nej: hush stdin MI_TOKEN -- npx wrangler secret put MI_TOKEN",
	"export": "uso: hush export [NOMBRES...]\nMuestra valores en TU terminal. Se niega si la salida no es una terminal (así ningún agente puede capturarlos).",
	"serve":  "uso: hush serve\nServidor MCP (stdio) para tu harness de IA. Registralo con: hush setup",
	"setup":  "uso: hush setup\nInstala el skill en tus harnesses (claude, codex, cursor, agents) y muestra cómo conectar el servidor MCP.",
	"update": "uso: hush update\nDescarga la última versión desde GitHub releases (verificada por checksum) y la instala.",
	"status": "uso: hush status\nMuestra versión, ubicación del vault y skills instalados.",
}

func help(argv []string, stdout io.Writer) int {
	if len(argv) == 0 {
		printUsage(stdout)
		return OK
	}
	if text, ok := helpText[argv[0]]; ok {
		fmt.Fprintln(stdout, text)
		return OK
	}
	fmt.Fprintf(stdout, "hush: no hay ayuda para %q\n\n", argv[0])
	printUsage(stdout)
	return Usage
}

func serve(store vault.Store, stderr io.Writer) int {
	if err := mcp.NewServer(store).ServeStdio(); err != nil {
		fmt.Fprintf(stderr, "hush: servidor MCP: %s\n", err)
		return Failed
	}
	return OK
}

func runSetup(stdout, stderr io.Writer) int {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(stderr, "hush: sin HOME")
		return Failed
	}
	results, err := setup.Apply(home)
	if err != nil {
		fmt.Fprintf(stderr, "hush: setup: %s\n", err)
		return Failed
	}
	for _, r := range results {
		fmt.Fprintf(stdout, "hush: skill %s → %s (%s)\n", r.Harness, r.Path, r.State)
	}
	fmt.Fprintln(stdout)
	fmt.Fprint(stdout, setup.Wiring())
	return OK
}

func runUpdate(stdout, stderr io.Writer) int {
	result, err := update.NewSelfUpdater().Run(context.Background(), Version)
	if err != nil {
		fmt.Fprintf(stderr, "hush: update: %s\n", err)
		return Failed
	}
	fmt.Fprintln(stdout, result.Message)
	if result.Updated {
		fmt.Fprintln(stdout, "Reiniciá tu harness para usar la nueva versión.")
	}
	return OK
}

func status(store vault.Store, stdout, stderr io.Writer) int {
	values, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, "hush: no se pudo leer el vault")
		return Failed
	}
	fmt.Fprintf(stdout, "hush %s\nvault: %s (%d secretos)\n", Version, store.Path(), len(values))
	home, err := os.UserHomeDir()
	if err != nil {
		return OK
	}
	ok := 0
	targets := setup.Targets(home)
	for _, t := range targets {
		if _, err := os.Stat(filepath.Join(t.Dir, "SKILL.md")); err == nil {
			ok++
		}
	}
	fmt.Fprintf(stdout, "skills: %d/%d harnesses\n", ok, len(targets))
	return OK
}

func exportSecrets(argv []string, store vault.Store, stdout, stderr io.Writer) int {
	if !canReveal(stdout) {
		fmt.Fprintln(stderr, "hush: export solo en terminal real (stdout no es TTY). Así ningún harness puede capturar valores.")
		return Failed
	}
	values, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, "hush: no se pudo leer el vault")
		return Failed
	}
	names := argv
	if len(names) == 0 {
		for k := range values {
			names = append(names, k)
		}
	}
	for _, n := range names {
		v, ok := values[n]
		if !ok {
			fmt.Fprintf(stderr, "hush: falta %s\n", n)
			return Missing
		}
		fmt.Fprintf(stdout, "%s=%s\n", n, v)
	}
	return OK
}

func canReveal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
