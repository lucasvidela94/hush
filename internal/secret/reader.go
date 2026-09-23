package secret

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func Read(stdin *os.File) (string, error) {
	fi, err := stdin.Stat()
	if err == nil && fi.Mode()&os.ModeCharDevice == 0 {
		raw, err := io.ReadAll(stdin)
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
	off := exec.Command("stty", "-echo")
	off.Stdin = tty
	_ = off.Run()
	defer func() {
		on := exec.Command("stty", "echo")
		on.Stdin = tty
		_ = on.Run()
		fmt.Fprintln(tty)
	}()
	line, _ := bufio.NewReader(tty).ReadString('\n')
	return strings.TrimRight(line, "\r\n"), nil
}
