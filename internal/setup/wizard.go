package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func IsTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func Wizard(home string, stdin *os.File, stdout io.Writer, command string) int {
	fmt.Fprintln(stdout, "hush setup — conecto el vault con tus harnesses.")
	fmt.Fprintln(stdout)
	results, err := Apply(home)
	if err != nil {
		fmt.Fprintf(stdout, "hush: skills: %s\n", err)
		return 1
	}
	for _, r := range results {
		fmt.Fprintf(stdout, "  skill %-7s %s\n", r.Harness, r.State)
	}

	pathEnv := os.Getenv("PATH")
	found := []Harness{}
	for _, h := range Harnesses() {
		if h.Detect(home, pathEnv) {
			found = append(found, h)
		}
	}
	if len(found) == 0 {
		fmt.Fprintln(stdout, "\nNo detecté ningún harness. Registro manual:")
		fmt.Fprint(stdout, Wiring())
		return 0
	}

	reader := bufio.NewReader(stdin)
	for _, h := range found {
		fmt.Fprintf(stdout, "\n## %s\n  %s\n", h.Name, h.Snippet(command))
		if h.Apply == nil {
			fmt.Fprintln(stdout, "  (agregalo a mano, no toco ese formato)")
			continue
		}
		if !ask(reader, stdout, "  ¿lo registro? [S/n] ") {
			continue
		}
		msg, err := h.Apply(home, command)
		if err != nil {
			fmt.Fprintf(stdout, "  ⚠ %s — hacelo manual con el snippet de arriba\n", err)
			continue
		}
		fmt.Fprintf(stdout, "  ✓ %s\n", msg)
	}
	fmt.Fprintln(stdout, "\nListo. Reiniciá tu harness para que tome el servidor MCP.")
	return 0
}

func ask(reader *bufio.Reader, stdout io.Writer, prompt string) bool {
	fmt.Fprint(stdout, prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "" || answer == "s" || answer == "si" || answer == "sí" || answer == "y" || answer == "yes"
}
