# hush — el agente usa secretos sin verlos

```bash
npm i -g hush-secrets          # instala binario + skill en tus harnesses
hush set MY_API_KEY            # en TU terminal, nunca en el chat
hush update                    # self-update desde GitHub releases
```

## Por qué

Pegar un secreto en el chat del harness lo deja en el transcript para siempre.
`hush` separa los canales: el humano carga por TTY o por el prompt nativo del
harness (vía MCP), el agente solo orquesta nombres. Ninguna tool devuelve valores.

## MCP (harness agnóstico: claude, opencode, cursor, codex)

`hush serve` expone stdio con 4 tools: `hush_check`, `hush_list`, `hush_need`
(pide el valor por elicitation y lo guarda), `hush_run` (inyecta + redacta).
`hush setup` instala el skill y muestra cómo registrar el servidor en cada harness.

## CLI

```bash
openssl rand -hex 24 | hush set WHATSAPP_VERIFY_TOKEN
hush check WHATSAPP_VERIFY_TOKEN META_APP_SECRET
hush stdin WHATSAPP_VERIFY_TOKEN -- npx wrangler secret put WHATSAPP_VERIFY_TOKEN
hush run --only META_APP_SECRET -- npx wrangler deploy
hush status
```

`check` y `list` muestran solo nombres. `run` y `stdin` inyectan en el
proceso hijo y redactan la salida a `[REDACTED]`.

## Desarrollo

```bash
make check   # generate + fmt + vet + test + build
make hook    # pre-commit (generate + fmt + vet + test)
./scripts/release.sh 0.2.0  # bump + tag (el workflow publica npm + release)
```

Commits convencionales: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`.
`SKILL.md` es fuente única del skill; `internal/setup/skill.md` se regenera con
`go generate ./...` (hay test que lo exige).
