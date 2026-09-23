package secret

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func Read(stdin *os.File) (string, error) {
	fi, err := stdin.Stat()
	if err == nil && fi.Mode()&os.ModeCharDevice == 0 {
		raw, err := io.ReadAll(stdin)
		return strings.TrimRight(string(raw), "\r\n"), err
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stderr, "valor (no se muestra): ")
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		return strings.TrimRight(string(raw), "\r\n"), err
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprint(os.Stderr, "valor: ")
		line, _ := bufio.NewReader(stdin).ReadString('\n')
		return strings.TrimRight(line, "\r\n"), nil
	}
	defer func() { _ = tty.Close() }()
	fmt.Fprint(tty, "valor (no se muestra): ")
	raw, err := term.ReadPassword(int(tty.Fd()))
	fmt.Fprintln(tty)
	return strings.TrimRight(string(raw), "\r\n"), err
}
