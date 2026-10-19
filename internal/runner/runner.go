package runner

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"hush/internal/redact"
)

const captureLimit = 1 << 20

const killGrace = 2 * time.Second

func Run(ctx context.Context, extra map[string]string, argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return run(ctx, extra, argv, stdin, stdout, stderr, false)
}

// RunStream ejecuta como Run pero reenvía stdout/stderr por línea a medida
// que llegan, con redacción incremental. Es para procesos longevos con stdio
// en streaming (servidores MCP): el modo batch retiene todo hasta el exit.
func RunStream(ctx context.Context, extra map[string]string, argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return run(ctx, extra, argv, stdin, stdout, stderr, true)
}

func run(ctx context.Context, extra map[string]string, argv []string, stdin io.Reader, stdout, stderr io.Writer, stream bool) int {
	env := os.Environ()
	secrets := make([]string, 0, len(extra))
	for k, v := range extra {
		env = append(env, k+"="+v)
		secrets = append(secrets, v)
	}
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()
	before := snapshot(watchedPaths(home, cwd))
	code := execute(ctx, env, secrets, argv, stdin, stdout, stderr, stream)
	warnPersisted(stderr, findPersisted(before, snapshot(watchedPaths(home, cwd)), extra))
	return code
}

func Pipe(ctx context.Context, value string, argv []string, stdout, stderr io.Writer) int {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return 1
	}
	go func() {
		_, _ = pipeWriter.WriteString(value + "\n")
		_ = pipeWriter.Close()
	}()
	return execute(ctx, os.Environ(), []string{value}, argv, pipeReader, stdout, stderr, false)
}

func execute(ctx context.Context, env []string, secrets []string, argv []string, stdin io.Reader, stdout, stderr io.Writer, stream bool) int {
	if len(argv) == 0 {
		return 1
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = killGrace
	var err error
	if stream {
		outS, errS := redact.NewStream(stdout, secrets), redact.NewStream(stderr, secrets)
		cmd.Stdout = outS
		cmd.Stderr = errS
		err = cmd.Run()
		_ = outS.Flush()
		_ = errS.Flush()
	} else {
		var outB, errB cappedBuffer
		cmd.Stdout = &outB
		cmd.Stderr = &errB
		err = cmd.Run()
		out := append(outB.truncatedNote(), redact.Apply(outB.Bytes(), secrets)...)
		_, _ = stdout.Write(out)
		_, _ = stderr.Write(redact.Apply(errB.Bytes(), secrets))
	}
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return 1
}

type cappedBuffer struct {
	buf       bytes.Buffer
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	room := captureLimit - b.buf.Len()
	if room <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if len(p) > room {
		p = p[:room]
		b.truncated = true
	}
	return b.buf.Write(p)
}

func (b *cappedBuffer) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *cappedBuffer) truncatedNote() []byte {
	if b.truncated {
		return []byte("\n[hush: salida recortada por límite]\n")
	}
	return nil
}
