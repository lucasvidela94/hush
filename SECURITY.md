# Security policy

## Threat model

`hush` keeps secret **values** out of AI agent context (chat transcripts, tool
outputs, logs). It does this by:

- Storing values in `~/.hush/vault` (file `0600`, dir `0700`).
- Injecting values only into child-process environments (`hush run`) or stdin
  (`hush stdin`), redacting them from captured output.
- Never returning values from any MCP tool (`check`/`list` return names only).
- Accepting new values only via local TTY, pipe, or MCP elicitation — never chat.

## What hush does NOT protect against

- **A compromised machine.** Anyone (or any process) running as your user can
  read `~/.hush/vault`, run `hush export` in a terminal, or attach to child
  processes. hush raises the bar against accidental leaks, it is not a vault
  against a local attacker. For that, use a system keychain or HSM.
- **Short secrets.** Redaction skips values under 4 characters to avoid
  mangling normal output. PINs and short codes can leak through `run` output.
- **Exfiltration by the child process.** A command run via `hush run` receives
  real values and can send them anywhere. Only run commands you trust.
- **Elicited values and the MCP client.** Values typed into a harness prompt
  travel through the client app (never the model in compliant clients).
  Maximum paranoia: provision with `hush set` in your own terminal.
- **Self-update supply chain.** `hush update` verifies the sha256 checksum,
  but checksum and binary travel the same channel (GitHub releases). A
  compromised release would verify cleanly. Pinned, signed releases are
  future work.
- **Concurrent writes.** Two simultaneous `set`/`need` calls can lose one
  value (last-writer-wins). Re-run `set` if a value goes missing.
- **Already-pasted secrets.** Nothing removes a value from a chat transcript.
  Rotate it at the source.

## Reporting

Open a GitHub issue at https://github.com/lucasvidela94/hush/issues.
Do not paste real secrets in reports — use `REDACTED` placeholders.
