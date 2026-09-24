package secret

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	stop := make(chan struct{})
	defer close(stop)
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(sigch)
	go func() {
		select {
		case <-sigch:
			_ = term.Restore(fd, oldState)
			os.Exit(130)
		case <-stop:
		}
	}()
	defer func() { _ = term.Restore(fd, oldState) }()
	fmt.Fprint(w, "valor: ")
	f := os.NewFile(uintptr(fd), "hush-tty")
	br := bufio.NewReader(f)
	var runes []rune
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			fmt.Fprintln(w)
			if err == io.EOF {
				return string(runes), nil
			}
			return string(runes), err
		}
		switch r {
		case '\r', '\n':
			fmt.Fprintln(w)
			return string(runes), nil
		case 3: // Ctrl-C
			_ = term.Restore(fd, oldState)
			fmt.Fprintln(w)
			os.Exit(130)
			return "", nil
		case 4: // Ctrl-D
			if len(runes) == 0 {
				fmt.Fprintln(w)
				return "", nil
			}
		case 127, 8: // Backspace/DEL
			if len(runes) > 0 {
				runes = runes[:len(runes)-1]
				fmt.Fprint(w, "\b \b")
			}
		case 21: // Ctrl-U: borra línea
			for range runes {
				fmt.Fprint(w, "\b \b")
			}
			runes = nil
		default:
			if r < 32 {
				continue
			}
			runes = append(runes, r)
			fmt.Fprint(w, "*")
		}
	}
}
