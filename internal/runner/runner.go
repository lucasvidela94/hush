package runner

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"

	"hush/internal/redact"
)

const captureLimit = 1 << 20

func Run(ctx context.Context, extra map[string]string, argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	env := os.Environ()
	secrets := make([]string, 0, len(extra))
	for k, v := range extra {
		env = append(env, k+"="+v)
		secrets = append(secrets, v)
	}
	return execute(ctx, env, secrets, argv, stdin, stdout, stderr)
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
	return execute(ctx, os.Environ(), []string{value}, argv, pipeReader, stdout, stderr)
}

func execute(ctx context.Context, env []string, secrets []string, argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(argv) == 0 {
		return 1
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = env
	cmd.Stdin = stdin
	var outB, errB cappedBuffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB
	err := cmd.Run()
	out := append(outB.truncatedNote(), redact.Apply(outB.Bytes(), secrets)...)
	_, _ = stdout.Write(out)
	_, _ = stderr.Write(redact.Apply(errB.Bytes(), secrets))
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
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
