package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"hush/internal/vault"
)

type fakeElicitor struct {
	action  mcp.ElicitationResponseAction
	content any
	err     error
}

func (f fakeElicitor) RequestElicitation(ctx context.Context, request mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &mcp.ElicitationResult{
		ElicitationResponse: mcp.ElicitationResponse{Action: f.action, Content: f.content},
	}, nil
}

func testServer(store vault.Store, elicitor elicitor) *Server {
	s := NewServer(store)
	if elicitor != nil {
		s.elicitor = elicitor
	}
	return s
}

func callRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if res.IsError {
		t.Fatalf("resultado es error: %+v", res.Content)
	}
	out := ""
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			out += tc.Text
		}
	}
	return out
}

func TestNeedPresente(t *testing.T) {
	store := vault.New(t.TempDir())
	if err := store.Save(map[string]string{"K": "v-largo"}); err != nil {
		t.Fatal(err)
	}
	s := testServer(store, nil)
	res, err := s.handleNeed(context.Background(), callRequest(map[string]any{"name": "K"}))
	if err != nil {
		t.Fatal(err)
	}
	if got := resultText(t, res); !strings.Contains(got, "ya está guardado") {
		t.Fatalf("inesperado: %q", got)
	}
}

func TestNeedAceptaYGuarda(t *testing.T) {
	store := vault.New(t.TempDir())
	s := testServer(store, fakeElicitor{
		action:  mcp.ElicitationResponseActionAccept,
		content: map[string]any{"value": "nuevo-secreto"},
	})
	res, err := s.handleNeed(context.Background(), callRequest(map[string]any{"name": "N", "hint": "dashboard"}))
	if err != nil {
		t.Fatal(err)
	}
	if got := resultText(t, res); !strings.Contains(got, "guardado") {
		t.Fatalf("inesperado: %q", got)
	}
	values, _ := store.Load()
	if values["N"] != "nuevo-secreto" {
		t.Fatalf("no se guardó: %v", values)
	}
}

func TestNeedRechazoCierraSinGuardar(t *testing.T) {
	store := vault.New(t.TempDir())
	s := testServer(store, fakeElicitor{action: mcp.ElicitationResponseActionDecline})
	res, err := s.handleNeed(context.Background(), callRequest(map[string]any{"name": "N"}))
	if err != nil {
		t.Fatal(err)
	}
	got := resultText(t, res)
	if !strings.Contains(got, "hush set N") {
		t.Fatalf("sin fallback a terminal: %q", got)
	}
	values, _ := store.Load()
	if len(values) != 0 {
		t.Fatalf("guardó sin aceptar: %v", values)
	}
}

func TestNeedSinSoporteInteractivo(t *testing.T) {
	store := vault.New(t.TempDir())
	s := testServer(store, fakeElicitor{err: errors.New("sin elicitation")})
	res, err := s.handleNeed(context.Background(), callRequest(map[string]any{"name": "N"}))
	if err != nil {
		t.Fatal(err)
	}
	if got := resultText(t, res); !strings.Contains(got, "hush set N") {
		t.Fatalf("sin fallback a terminal: %q", got)
	}
}

func TestRunRedacta(t *testing.T) {
	store := vault.New(t.TempDir())
	if err := store.Save(map[string]string{"T": "token-ultrasecreto"}); err != nil {
		t.Fatal(err)
	}
	s := testServer(store, nil)
	res, err := s.handleRun(context.Background(), callRequest(map[string]any{
		"command": []any{"sh", "-c", "echo $T"},
		"only":    []any{"T"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	got := resultText(t, res)
	if strings.Contains(got, "token-ultrasecreto") {
		t.Fatalf("valor visible: %q", got)
	}
	if !strings.Contains(got, "exit=0") || !strings.Contains(got, "[REDACTED]") {
		t.Fatalf("incompleto: %q", got)
	}
}

func TestRunStdinFaltante(t *testing.T) {
	store := vault.New(t.TempDir())
	s := testServer(store, nil)
	res, err := s.handleRun(context.Background(), callRequest(map[string]any{
		"command":    []any{"cat"},
		"stdin_name": "AUSENTE",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got := resultText(t, res); !strings.Contains(got, "hush set AUSENTE") {
		t.Fatalf("sin fallback: %q", got)
	}
}

func TestCheckSoloNombres(t *testing.T) {
	store := vault.New(t.TempDir())
	if err := store.Save(map[string]string{"A": "valor-a-largo"}); err != nil {
		t.Fatal(err)
	}
	s := testServer(store, nil)
	res, err := s.handleCheck(context.Background(), callRequest(map[string]any{
		"names": []any{"A", "B"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	got := resultText(t, res)
	if strings.Contains(got, "valor-a-largo") {
		t.Fatalf("valor visible: %q", got)
	}
	if !strings.Contains(got, `"missing":["B"]`) {
		t.Fatalf("incompleto: %q", got)
	}
}
