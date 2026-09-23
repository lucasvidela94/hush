# Roadmap: from transcript hygiene to real boundary

`hush` today protects the **transcript**. A hostile or prompt-injected agent
with shell access can still exfiltrate, because any use-oracle is abusable:
if the agent can cause secret *use*, it can leak the *value*. Storage
tricks alone (keyring, encryption, daemon) do not close that hole for a
same-user agent. The boundary has to be **per-use human authorization** plus
**smaller blast radius**. That is what this roadmap builds.

## v1.0 — policies and audit

- **Per-secret policy file** (`~/.hush/policy`): which names exist, which
  commands may use them (`hush run` outside the list asks even in CLI),
  `output: opaque` flag per secret (exit code + stderr tail only, no stdout).
- **Append-only usage log** (`hush log`): timestamp, tool, secret names,
  command shape — never values. Detect misuse after the fact.
- **Optional keyring backend** (macOS Keychain / libsecret): defense in
  depth where the OS enforces something; file vault stays the default for
  servers and portability.
- **Session approvals**: `hush allow <names> --for 30m` in your terminal so
  the MCP confirm gate can say yes without a prompt per call, bound in time
  and scope.

## v1.1 — confinement

- **Separate-UID install guide**: daemon user + group socket for setups where
  the agent runs as a different user (CI runners, shared boxes). Documented
  script, not default — same-user desktops gain nothing from it.
- **Egress posture**: document running secret-using commands without network
  (`unshare -n`, containers) for the truly paranoid flows.

## Non-goals

- Blocking a same-user hostile agent from *using* approved secrets. Impossible
  without OS isolation; we say so instead of selling it.
- Becoming a team secrets manager (Vault, Doppler, 1Password own that).
- Windows support (TTY model).
