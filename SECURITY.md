# Security policy

## Threat model

`hush` protects the **transcript**: secret values stay out of AI agent
context (chat, tool outputs, logs) through its own channels. It does **not**
protect against a hostile or prompt-injected agent with shell access. Such an
agent can read `~/.hush/vault` directly, persist stdin to disk (`tee`), or
transform values (base64, reverse, hex) past literal redaction. Redaction of
common encodings, deny-by-default injection, and the TTY-only export raise
the cost of accidents, not of malice. If your adversary is the agent itself,
you need OS-level isolation (keychain, daemon + socket auth), not hush.

Concretely, hush:

- Stores values in `~/.hush/vault` (file `0600`, dir `0700`, atomic write,
  symlinks refused).
- Injects values only into child-process environments (`hush run`) or stdin
  (`hush stdin`), redacting literal values plus base64/hex/URL/reversed forms
  from captured output (bounded at 1 MB per stream).
- Never returns values from any MCP tool (`check`/`list` return names only).
- Injects nothing unless `only[]`/`--only` or `all`/`--all` is explicit.
- Accepts new values only via local TTY, pipe, or MCP elicitation — never chat.

## What hush does NOT protect against

- **A compromised machine.** Anyone (or any process) running as your user can
  read `~/.hush/vault`, run `hush export` in a terminal, or attach to child
  processes. hush raises the bar against accidental leaks, it is not a vault
  against a local attacker. For that, use a system keychain or HSM.
- **Short secrets.** Redaction skips values under 4 characters to avoid
  mangling normal output. PINs and short codes can leak through `run` output.
- **Split transforms.** Redaction covers literal values plus base64, hex
  (plain/upper/spaced), URL-encoding, and reversal. Anything that splits the
  value with whitespace, wraps base64 lines, or re-encodes it (gzip,
  encryption) passes through. Enumeration is finite; exfiltration is not.
- **Exfiltration by the child process.** A command run via `hush run` receives
  real values and can send them anywhere. Only run commands you trust.
- **Elicited values and the MCP client.** Values typed into a harness prompt
  travel through the client app (never the model in compliant clients).
  Maximum paranoia: provision with `hush set` in your own terminal.
- **Self-update supply chain.** `hush update` exige `checksums.txt.sig`
  firmado por una clave embebida (lista para rotación). Sin firma válida no
  instala nada. La clave privada vive en un secret de CI y los releases pasan
  por environment con aprobación. Modelo completo abajo.
- **Concurrent writes.** Protegidos con `flock` en `set`/`need`.
- **Already-pasted secrets.** Nothing removes a value from a chat transcript.
  Rotate it at the source.

## Shell model: what the MCP confirm does and does not govern

`hush_run` confirmation governs calls **through MCP**. It does not govern an
agent with shell access: `hush run --all -- …`, `cat ~/.hush/vault`, or
`hush export` under a pty run without asking. The real control there is the
harness's own permissions. Recommended (verify syntax against your harness
version, it changes):

- Claude Code (`settings.json`): deny `Read(~/.hush/**)`, ask on
  `Bash(hush run:*)` and `Bash(hush export:*)`.
- Equivalent deny/ask rules exist in most harnesses for shell and file tools.

This is whack-a-mole against `cat` via Bash — useful friction, not a
boundary. The structural fix is getting the vault out of the agent's reach
(system keychain, or a daemon + socket auth where the agent's process cannot
read the file). That is the roadmap; hush today protects the transcript.

## Output policy: why not opaque-by-default

Every `hush_run` execution was approved by a human (command + secret names
via prompt). Making output opaque by default would not stop an injected
agent — it already passed a human gate — but would break legitimate flows
(deploy confirmations, `wrangler` success messages). So: full output with
layered redaction (literals, common encodings, fail-closed transform
detection), bounded at 32 KB per call. Halves of a secret exfiltrated across
separate calls cannot be detected statelessly; no new one-off encodings will
be added — only generic transforms.

## Signing and rotation model

- La clave privada de firmas solo existe en dos lugares: offline con el
  maintainer y en el secret `HUSH_SIGN_PRIV` de CI. CI firma automáticamente
  cada release; sin secreto, el release falla (fail-closed).
- `internal/update/pubkey.go` embebe la lista de claves públicas aceptadas.
- Rotación: agregar la nueva clave a la lista, publicar un release (migra a
  todos los que actualizan), y después rotar el secret. Nunca reutilizar una
  clave expuesta en un transcript.
- Binarios v0.4.1 llevan la clave vieja (privada perdida): no verifican
  firmas nuevas. Reinstalar por npm (`npm i -g hush-secrets`).

## Reporting

Open a GitHub issue at https://github.com/lucasvidela94/hush/issues.
Do not paste real secrets in reports — use `REDACTED` placeholders.
