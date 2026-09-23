# hush — AI agents use secrets without ever seeing them

## Scope honesty

`hush` protects the **transcript**: values never enter chat, tool outputs, or
logs through its own channels. It does **not** protect against a hostile or
prompt-injected agent with shell access — such an agent can `cat` the vault,
`tee` stdin to disk, or transform values (base64/reverse) past literal
redaction. If your adversary is the agent itself, you need OS-level isolation
(keychain, daemon + socket auth), not hush. See `SECURITY.md`.

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

## Where secrets live, and getting them back

Values are stored in `~/.hush/vault` (file `0600`, dir `0700`), one
`NAME=value` per line. There is intentionally **no** read-back over MCP and
no `get` command — anything an agent can call ends up in its context.

To recover a value (rotation, migration, backup), use a real terminal:

```bash
hush export                  # prints all, NAME=value per line
hush export META_APP_SECRET  # prints one
```

`export` refuses to run unless stdout is a TTY, so a harness cannot capture it
by piping (a pty is not proof of a human — see `SECURITY.md`). Back up the file itself (`cp ~/.hush/vault …`) for disaster
recovery; wipe with `rm -rf ~/.hush`. See `SECURITY.md` and `PRIVACY.md`.

## Notes

- Linux and macOS only (amd64/arm64). No Windows (TTY model).
- `hush_run` over MCP asks for human confirmation on every call and times
  out slow commands at 5 minutes. Long-lived processes belong in your
  terminal, not in the agent.
- Elicitation (native secret prompts) verified on Claude Code; other
  harnesses fall back to `hush set` in your terminal. `hush doctor` tells you
  where you stand.

## Docs

`SECURITY.md` (threat model) · `PRIVACY.md` · `CONTRIBUTING.md` · `CHANGELOG.md`

## Alternatives

| | hush | psst | key-amnesia | akm | 1Password MCP |
|---|---|---|---|---|---|
| Values hidden from transcript | ✅ | ✅ | ✅ | ✅ | ✅ |
| Hostile-agent isolation | ❌ honest | ❌ | parcial | parcial | ❌ |
| MCP server | ✅ | ❌ | ❌ | ❌ | ✅ (Codex only) |
| Load via native prompt (no chat) | ✅ elicitation | ❌ | ✅ popup | ❌ | ✅ |
| TTY-only export | ✅ | ❌ | ❌ | ❌ | ❌ |
| Vault corruption guard | ✅ | n/a | ❌ | ❌ | n/a |
| Harness-agnostic setup | ✅ | ❌ | parcial | ❌ | ❌ |
| Single binary, no runtime | ✅ Go | ✅ | ❌ Python | ✅ Rust* | ❌ |
| Works offline / local-only | ✅ | ✅ | ✅ | ✅ | ❌ |

*akm is macOS-only (Keychain). hush keeps a plain `0600` file so any dev can
inspect, back up, and recover it with standard tools.

## Development

```bash
make check   # generate + fmt + vet + test + build
make hook    # pre-commit (generate + fmt + vet + test)
./scripts/release.sh 0.2.0  # bump + tag (workflow publishes npm + release)
```

Conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`.
`SKILL.md` is the single source of the skill; `internal/setup/skill.md` is
regenerated with `go generate ./...` (a test enforces it).
