package runner

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"hush/internal/redact"
)

func Run(extra map[string]string, argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	env := os.Environ()
	secrets := make([]string, 0, len(extra))
	for k, v := range extra {
		env = append(env, k+"="+v)
		secrets = append(secrets, v)
	}
	return execute(env, secrets, argv, stdin, stdout, stderr)
}

func Pipe(value string, argv []string, stdout, stderr io.Writer) int {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return 1
	}
	go func() {
		_, _ = pipeWriter.WriteString(value + "\n")
		_ = pipeWriter.Close()
	}()
	return execute(os.Environ(), []string{value}, argv, pipeReader, stdout, stderr)
}

func execute(env []string, secrets []string, argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = env
	cmd.Stdin = stdin
	var outB, errB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB
	err := cmd.Run()
	_, _ = stdout.Write(redact.Apply(outB.Bytes(), secrets))
	_, _ = stderr.Write(redact.Apply(errB.Bytes(), secrets))
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	return 1
}
