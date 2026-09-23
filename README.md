# hush — AI agents use secrets without ever seeing them

```bash
npm i -g --allow-scripts=hush-secrets hush-secrets  # installs binary + skill
hush set MY_API_KEY            # in YOUR terminal, never in chat
hush update                    # self-update from GitHub releases
```

## Why

Pasting a secret into a harness chat leaves it in the transcript forever.
`hush` separates the channels: the human loads values via TTY or the harness
native prompt (over MCP), the agent only orchestrates names. No tool ever
returns a value.

## MCP (harness-agnostic: claude, opencode, cursor, codex)

`hush serve` exposes stdio with 4 tools: `hush_check`, `hush_list`, `hush_need`
(asks for the value via elicitation and stores it), `hush_run` (injects +
redacts). `hush setup` installs the skill and shows how to register the
server in each harness.

## CLI

```bash
openssl rand -hex 24 | hush set WHATSAPP_VERIFY_TOKEN
hush check WHATSAPP_VERIFY_TOKEN META_APP_SECRET
hush stdin WHATSAPP_VERIFY_TOKEN -- npx wrangler secret put WHATSAPP_VERIFY_TOKEN
hush run --only META_APP_SECRET -- npx wrangler deploy
hush status
```

`check` and `list` print names only. `run` and `stdin` inject into the child
process and redact output to `[REDACTED]`.

## Development

```bash
make check   # generate + fmt + vet + test + build
make hook    # pre-commit (generate + fmt + vet + test)
./scripts/release.sh 0.2.0  # bump + tag (workflow publishes npm + release)
```

Conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`.
`SKILL.md` is the single source of the skill; `internal/setup/skill.md` is
regenerated with `go generate ./...` (a test enforces it).
