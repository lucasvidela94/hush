# hush

CLI y servidor MCP en Go para usar secretos sin que el agente los vea.

## Comandos

- `make check`: generate + fmt + vet + lint + test -race + build. Obligatorio antes de dar algo por terminado.
- `make lint`: golangci-lint (config estricta en `.golangci.yml`).
- `go test -race ./internal/<pkg>`: un paquete.

## Decisiones

- `cmd/` son entrypoints finos; la lógica vive en `internal/<pkg>`.
- `SKILL.md` es la única fuente de la skill; `go generate ./...` refresca la copia embebida.
- Errores al usuario en español, identificadores en inglés. Errores envueltos con `%w`.
- Cambios en `redact`, `runner`, `secret`, `vault`, `mcp/handlers.go`, `install.js` o `update` exigen un test que pruebe que no se filtra ningún valor.

## Prohibido

- Paquetes `util/` o `common/`.
- Interfaces con una sola implementación.
- `panic`, variables globales mutables, `init()`.
- `//nolint` sin motivo escrito al lado.
- Comentarios que repiten el código.
- Nuevas dependencias sin pedirlo.

## Tests

- Table-driven, `t.TempDir()`, filesystem real en vez de mocks.
- Sin `time.Sleep`.
- Un bug se arregla con un test que falla primero.

## Flujo

Una rama y un PR por issue (`Closes #N`). Commits convencionales, sin atribución de IA.
