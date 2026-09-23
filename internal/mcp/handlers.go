package mcp

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"hush/internal/runner"
	"hush/internal/vault"
)

const maxOutput = 32 * 1024

const maxHint = 200

var validName = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

func stringArg(args map[string]any, name string) (string, error) {
	v, ok := args[name]
	if !ok {
		return "", fmt.Errorf("missing required parameter: %s", name)
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("parameter %s must be a non-empty string", name)
	}
	return s, nil
}

func stringList(args map[string]any, name string) ([]string, error) {
	v, ok := args[name]
	if !ok {
		return nil, nil
	}
	raw, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("parameter %s must be an array", name)
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("parameter %s must be an array of strings", name)
		}
		out = append(out, s)
	}
	return out, nil
}

func truncate(s string) string {
	if len(s) <= maxOutput {
		return s
	}
	return s[:maxOutput] + "\n[hush: salida truncada]"
}

func (s *Server) handleCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	names, err := stringList(request.GetArguments(), "names")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	values, err := s.store.Load()
	if err != nil {
		return mcp.NewToolResultError("hush: no se pudo leer el vault"), nil
	}
	missing := []string{}
	present := []string{}
	for _, n := range names {
		if _, ok := values[n]; ok {
			present = append(present, n)
		} else {
			missing = append(missing, n)
		}
	}
	return mcp.NewToolResultJSON(map[string][]string{"missing": missing, "present": present})
}

func (s *Server) handleList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	values, err := s.store.Load()
	if err != nil {
		return mcp.NewToolResultError("hush: no se pudo leer el vault"), nil
	}
	names := make([]string, 0, len(values))
	for k := range values {
		names = append(names, k)
	}
	if len(names) == 0 {
		return mcp.NewToolResultText("hush: vault vacío"), nil
	}
	return mcp.NewToolResultText(strings.Join(names, "\n")), nil
}

func (s *Server) handleNeed(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	name, err := stringArg(args, "name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if !validName.MatchString(name) {
		return mcp.NewToolResultError("name debe ser [A-Z_][A-Z0-9_]*"), nil
	}
	values, err := s.store.Load()
	if err != nil {
		return mcp.NewToolResultError("hush: no se pudo leer el vault"), nil
	}
	if _, ok := values[name]; ok {
		return mcp.NewToolResultText(fmt.Sprintf("hush: %s ya está guardado", name)), nil
	}

	hint, _ := args["hint"].(string)
	hint = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, hint)
	if len([]rune(hint)) > maxHint {
		hint = string([]rune(hint)[:maxHint])
	}
	message := fmt.Sprintf("Valor para %s. Se guarda localmente y nunca entra al chat.", name)
	if strings.TrimSpace(hint) != "" {
		message += " Dónde encontrarlo: " + strings.TrimSpace(hint)
	}

	value, ok := s.elicitValue(ctx, message)
	if !ok {
		return mcp.NewToolResultText(fallbackMessage(name)), nil
	}

	lock, err := vault.Acquire(s.store.Dir())
	if err != nil {
		return mcp.NewToolResultError("hush: no se pudo bloquear el vault"), nil
	}
	defer lock.Release()
	values, err = s.store.Load()
	if err != nil {
		return mcp.NewToolResultError("hush: no se pudo leer el vault"), nil
	}
	if _, ok := values[name]; ok {
		return mcp.NewToolResultText(fmt.Sprintf("hush: %s ya está guardado", name)), nil
	}
	values[name] = value
	if err := s.store.Save(values); err != nil {
		return mcp.NewToolResultError("hush: no se pudo guardar"), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("hush: %s guardado", name)), nil
}

func (s *Server) confirmRun(ctx context.Context, command []string, names []string) (bool, string) {
	elicitCtx, cancel := context.WithTimeout(ctx, s.elicitTimeout)
	defer cancel()
	req := mcp.ElicitationRequest{
		Params: mcp.ElicitationParams{
			Message: fmt.Sprintf("Ejecutar %s con secretos [%s]?", formatCommand(command), strings.Join(names, ", ")),
			RequestedSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"confirm": map[string]any{"type": "boolean", "description": "Confirmar ejecución"},
				},
				"required": []string{"confirm"},
			},
		},
	}
	result, err := s.elicitor.RequestElicitation(elicitCtx, req)
	if err != nil {
		return false, "hush: este cliente no soporta confirmación. Corré el comando en tu terminal con hush run."
	}
	if result.Action != mcp.ElicitationResponseActionAccept {
		return false, "hush: ejecución no confirmada por el humano."
	}
	content, _ := result.Content.(map[string]any)
	if confirm, _ := content["confirm"].(bool); confirm {
		return true, ""
	}
	return false, "hush: ejecución no confirmada por el humano."
}

func formatCommand(command []string) string {
	quoted := make([]string, 0, len(command))
	shown := 0
	for _, arg := range command {
		q := fmt.Sprintf("%q", arg)
		if shown+len(q) > 160 {
			return strings.Join(quoted, " ") + fmt.Sprintf(" (+%d caracteres)", len(strings.Join(command, " "))-shown)
		}
		quoted = append(quoted, q)
		shown += len(q) + 1
	}
	return strings.Join(quoted, " ")
}

func (s *Server) elicitValue(ctx context.Context, message string) (string, bool) {
	elicitCtx, cancel := context.WithTimeout(ctx, s.elicitTimeout)
	defer cancel()

	req := mcp.ElicitationRequest{
		Params: mcp.ElicitationParams{
			Message: message,
			RequestedSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"value": map[string]any{"type": "string", "description": "Secret value"},
				},
				"required": []string{"value"},
			},
		},
	}
	result, err := s.elicitor.RequestElicitation(elicitCtx, req)
	if err != nil || result.Action != mcp.ElicitationResponseActionAccept {
		return "", false
	}
	content, _ := result.Content.(map[string]any)
	value, _ := content["value"].(string)
	if strings.TrimSpace(value) == "" {
		return "", false
	}
	return strings.TrimSpace(value), true
}

func fallbackMessage(name string) string {
	return fmt.Sprintf("hush: sin valor para %s. Decile al humano que en su propia terminal corra: hush set %s. Nunca pidas el valor por chat.", name, name)
}

func (s *Server) handleRun(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	args := request.GetArguments()
	command, err := stringList(args, "command")
	if err != nil || len(command) == 0 {
		return mcp.NewToolResultError("parameter command must be a non-empty array of strings"), nil
	}
	only, err := stringList(args, "only")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	stdinName, _ := args["stdin_name"].(string)
	all, _ := args["all"].(bool)

	values, err := s.store.Load()
	if err != nil {
		return mcp.NewToolResultError("hush: no se pudo leer el vault"), nil
	}

	restrict := map[string]bool{}
	for _, k := range only {
		restrict[k] = true
	}
	extra := map[string]string{}
	for k, v := range values {
		if all || restrict[k] {
			extra[k] = v
		}
	}
	if len(extra) == 0 && strings.TrimSpace(stdinName) == "" {
		return mcp.NewToolResultError("hush_run exige only[] o all=true explícito (no inyecto todo por defecto)"), nil
	}

	names := []string{}
	for k := range extra {
		names = append(names, k)
	}
	stdinName = strings.TrimSpace(stdinName)
	if stdinName != "" {
		if _, ok := values[stdinName]; !ok {
			return mcp.NewToolResultText(fallbackMessage(stdinName)), nil
		}
		names = []string{stdinName}
	}
	sort.Strings(names)
	confirmed, message := s.confirmRun(ctx, command, names)
	if !confirmed {
		return mcp.NewToolResultText(message), nil
	}

	var out, errOut limitedBuffer
	var code int
	if stdinName != "" {
		code = runner.Pipe(ctx, values[stdinName], command, &out, &errOut)
	} else {
		code = runner.Run(ctx, extra, command, nil, &out, &errOut)
	}

	text := fmt.Sprintf("exit=%d\n%s", code, truncate(out.String()+errOut.String()))
	return mcp.NewToolResultText(text), nil
}

type limitedBuffer struct {
	buf strings.Builder
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	room := 2*maxOutput - b.buf.Len()
	if room <= 0 {
		return len(p), nil
	}
	if len(p) > room {
		p = p[:room]
	}
	return b.buf.Write(p)
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}
