package secret

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
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
		return readPassword(int(os.Stdin.Fd()), os.Stderr)
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprint(os.Stderr, "valor: ")
		line, _ := bufio.NewReader(stdin).ReadString('\n')
		return strings.TrimRight(line, "\r\n"), nil
	}
	defer func() { _ = tty.Close() }()
	return readPassword(int(tty.Fd()), tty)
}

func readPassword(fd int, w *os.File) (string, error) {
	state, err := term.GetState(fd)
	if err != nil {
		return "", err
	}
	stop := make(chan struct{})
	defer close(stop)
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	defer signal.Stop(sigch)
	go func() {
		select {
		case <-sigch:
			_ = term.Restore(fd, state)
			os.Exit(130)
		case <-stop:
		}
	}()
	fmt.Fprint(w, "valor (no se muestra): ")
	raw, err := term.ReadPassword(fd)
	fmt.Fprintln(w)
	return strings.TrimRight(string(raw), "\r\n"), err
}
