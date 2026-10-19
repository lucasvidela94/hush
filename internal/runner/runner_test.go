package runner

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRunRedactaEnv(t *testing.T) {
	var out, errB bytes.Buffer
	code := Run(
		context.Background(),
		map[string]string{"HUSH_T": "valor-ultrasecreto"},
		[]string{"sh", "-c", "echo la key es $HUSH_T"},
		nil, &out, &errB,
	)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(out.String(), "valor-ultrasecreto") {
		t.Fatalf("salida sin redactar: %q", out.String())
	}
	if !strings.Contains(out.String(), "[REDACTED]") {
		t.Fatalf("falta marca: %q", out.String())
	}
}

func TestPipeRedactaEco(t *testing.T) {
	var out, errB bytes.Buffer
	code := Pipe(context.Background(), "valor-ultrasecreto", []string{"cat"}, &out, &errB)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(out.String(), "valor-ultrasecreto") {
		t.Fatalf("salida sin redactar: %q", out.String())
	}
}

func TestRunPropagaExit(t *testing.T) {
	var out, errB bytes.Buffer
	code := Run(context.Background(), nil, []string{"sh", "-c", "exit 7"}, nil, &out, &errB)
	if code != 7 {
		t.Fatalf("exit %d, quiero 7", code)
	}
}

func TestMataGrupoCompleto(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	var out, errB bytes.Buffer
	start := time.Now()
	code := Run(ctx, nil, []string{"sh", "-c", "sleep 30 & sleep 30"}, nil, &out, &errB)
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("tardó %v, el grupo sobrevivió", elapsed)
	}
	if code == 0 {
		t.Fatal("exit 0 con kill")
	}
}

func TestRunStreamEntregaAntesDeEOF(t *testing.T) {
	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	var errB bytes.Buffer
	done := make(chan int, 1)
	go func() {
		done <- RunStream(
			context.Background(),
			nil,
			[]string{"sh", "-c", "echo hola; cat > /dev/null"},
			inR, outW, &errB,
		)
	}()

	lineCh := make(chan string, 1)
	go func() {
		var first bytes.Buffer
		tmp := make([]byte, 1)
		for {
			n, rerr := outR.Read(tmp)
			if n > 0 {
				first.Write(tmp[:n])
				if tmp[0] == '\n' {
					break
				}
			}
			if rerr != nil {
				break
			}
		}
		lineCh <- first.String()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	select {
	case line := <-lineCh:
		if !strings.Contains(line, "hola") {
			t.Fatalf("primera línea inesperada: %q", line)
		}
	case <-ctx.Done():
		t.Fatal("stream no entregó nada antes de EOF: retiene stdout")
	}

	_ = inW.Close()
	_ = outW.Close()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("exit %d, quiero 0", code)
		}
	case <-ctx.Done():
		t.Fatal("el hijo no terminó tras EOF")
	}
}

func TestRunStreamRedacta(t *testing.T) {
	cases := []struct {
		name   string
		script string
	}{
		{"literal", `echo la key es $HUSH_T`},
		{"partido entre escrituras", `printf 'la key es %.9s' "$HUSH_T"; printf '%s\n' "${HUSH_T#?????????}"`},
		{"un carácter por línea", `printf '%s' "$HUSH_T" | fold -w1`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errB bytes.Buffer
			code := RunStream(
				context.Background(),
				map[string]string{"HUSH_T": "valor-ultrasecreto-123"},
				[]string{"sh", "-c", tc.script},
				nil, &out, &errB,
			)
			if code != 0 {
				t.Fatalf("exit %d: %s", code, errB.String())
			}
			flat := strings.Join(strings.Fields(out.String()), "")
			for _, half := range []string{"valor-ultr", "asecreto-123"} {
				if strings.Contains(flat, half) {
					t.Fatalf("quedó visible %q en %q", half, out.String())
				}
			}
			if !strings.Contains(out.String(), "REDACTED") {
				t.Fatalf("falta marca: %q", out.String())
			}
		})
	}
}
