# Changelog

## v0.1.1

- Fix: drop `bin` from package.json (binary is installed by `install.js`).
- Docs: English README, `--allow-scripts` install flag.
- CI: OIDC-primary npm publish with token fallback, `workflow_dispatch`.

## v0.1.0

- Vault (`~/.hush/vault`, `0600`), TTY/pipe `set`, names-only `check`/`list`.
- `run`/`stdin` injection with `[REDACTED]` output filtering.
- MCP server (`serve`): `hush_check`, `hush_list`, `hush_need` (elicitation),
  `hush_run`.
- `setup` deploys the skill to claude/codex/cursor/agents harnesses.
- `update` self-update from GitHub releases (sha256-verified).
- npm package `hush-secrets` with bundled platform binaries.
