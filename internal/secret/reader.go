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
		return strings.TrimSpace(string(raw)), err
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprint(os.Stderr, "valor: ")
		line, _ := bufio.NewReader(stdin).ReadString('\n')
		return strings.TrimSpace(line), nil
	}
	defer func() { _ = tty.Close() }()
	fmt.Fprint(tty, "valor (no se muestra): ")
	off := exec.Command("stty", "-echo")
	off.Stdin = tty
	_ = off.Run()
	line, _ := bufio.NewReader(tty).ReadString('\n')
	on := exec.Command("stty", "echo")
	on.Stdin = tty
	_ = on.Run()
	fmt.Fprintln(tty)
	return strings.TrimSpace(line), nil
}
