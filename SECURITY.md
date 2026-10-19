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
  (plain/upper/spaced), URL-encoding, and reversal. It also catches
  substrings of 6+ characters (at least 40% of the value), reordered characters
  (`fold | sort`), and masked copies (`sed 's/[0-9]/#/g'`); expect false
  positives on short secrets. Anything that splits the
  value with whitespace, wraps base64 lines, or re-encodes it (gzip,
  encryption) passes through. Enumeration is finite; exfiltration is not.
- **Exfiltration by the child process.** A command run via `hush run` receives
  real values and can send them anywhere. Only run commands you trust.
- **Values persisted by the child.** If a command run via `hush run` writes
  the value to disk (an MCP config, `.env`, a git remote URL), redaction does
  not cover later commands that read it back outside hush (`claude mcp get`,
  `printenv`, `cat`). Verify with commands that do not print the config, read
  back through `hush run`, and prefer env-var references over stored values.
  As a safety net, `hush run` warns on stderr (secret name and path, never the
  value) when a fixed set of files changed during the run and now contain an
  injected value literally: `~/.claude.json`, `~/.cursor/mcp.json`,
  `~/.config/opencode/opencode.json`, and `.mcp.json`, `opencode.json` and
  `.env*` in the current directory. It does not see transformed or encrypted
  copies, files outside that list, or a rewrite that keeps the same size and
  modification time.
- **Confirmation differs by path.** `hush_run` over MCP always asks a human to
  confirm; `hush run` from a shell only goes through the harness's normal
  permission prompt.
- **Elicited values and the MCP client.** Values typed into a harness prompt
  travel through the client app (never the model in compliant clients).
  Maximum paranoia: provision with `hush set` in your own terminal.
- **Self-update supply chain.** `hush update` no descarga ni verifica nada por
  su cuenta: consulta la última versión a npm y corre `npm install -g
  hush-secrets@VERSION`. La confianza es la del registro de npm y la
  publicación desde CI con OIDC (Trusted Publishing); hush no embebe ni rota
  claves. Modelo completo abajo.
- **Concurrent writes.** Protegidos con `flock` en `set`/`need`.
- **Already-pasted secrets.** Nothing removes a value from a chat transcript.
  Rotate it at the source.

## Shell model: what the MCP confirm does and does not govern

`hush_run` confirmation governs calls **through MCP**. It does not govern an
agent with shell access: `hush run --all -- …`, `cat ~/.hush/vault`, or
`hush export` under a pty run without asking. The real control there is the
harness's own permissions. `hush setup` offers to write them (and
`hush doctor` warns when they are missing), merging into existing rules with
a backup:

- Claude Code (`~/.claude/settings.json`): deny `Read(~/.hush/**)` and
  `Edit(~/.hush/**)` (they also cover `cat`, `head`, `sed` and redirections),
  ask on `Bash(hush run *)`, `Bash(hush stdin *)` and `Bash(hush export *)`.
- opencode v2 (`permissions`, last match wins, so they go at the end): deny
  `external_directory`, `read` and `edit` on `~/.hush/*`, deny `shell` on
  `*.hush/*`, ask on `shell` `hush run *`, `hush stdin *`, `hush export *`.
  A v1 config (`permission`) is left untouched: add the equivalents by hand.
  Agents with their own `permissions` load after the global rules and win:
  an agent with `read ** allow` or `shell grep * allow` overrides them.
  Repeat the hush rules at the end of that agent's list.
- Codex: `hush setup` writes `~/.codex/rules/hush.rules` (its own file) with
  `prompt` on `hush run`, `stdin` and `export`. Codex loads every `*.rules`
  and the strictest decision wins, so an `allow` approved earlier in
  `default.rules` does not override it. Codex can also deny `~/.hush` in its
  OS sandbox, which blocks `cat`, `grep` and any shell command, unlike tool
  rules. `hush setup` does not edit `config.toml`; `hush doctor` prints the
  permission profile to add. With it, `hush` run through the shell cannot
  read the vault either: use the hush MCP server (Codex launches MCP servers
  outside the sandbox) or approve running outside the sandbox.
- Other harnesses: equivalent deny/ask rules exist for shell and file tools;
  add them by hand.

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

## Distribution and update model

- El único canal de instalación es el paquete npm `hush-secrets`, publicado
  desde el workflow de release con OIDC (sin token de larga vida). Los
  binarios viajan dentro del paquete.
- `hush update` es `npm install -g`. No hay claves de firma que custodiar ni
  rotar: no existe un caso en que una clave perdida deje a los usuarios sin
  actualizar.
- Los releases de GitHub siguen publicando los binarios con `checksums.txt` y
  atestación de procedencia de build, útiles para auditar a mano.
- Hasta v0.6.0, `hush update` verificaba una firma ed25519 con claves
  embebidas. Esas versiones deben reinstalar una vez con `npm i -g
  hush-secrets@latest`.

## Reporting

Open a GitHub issue at https://github.com/lucasvidela94/hush/issues.
Do not paste real secrets in reports — use `REDACTED` placeholders.
