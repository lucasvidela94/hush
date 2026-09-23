# Changelog

## v0.3.0

- `run` deny-by-default: exige `--only`/`--all` (CLI) y `only[]`/`all` (MCP).
- Redacción endurecida: base64/hex/URL/reversa, orden por longitud, captura
  acotada a 1 MB, cancelación por contexto.
- Vault atómico (tmp+fsync+rename), symlinks rechazados, permisos reforzados,
  multilínea con escapes (PEM y JSON ya entran).
- `hush_need` valida nombres (`[A-Z_][A-Z0-9_]*`) y acota el hint.
- Updater testeado end-to-end (servidor local) y checksums sin prefijo.
- CI: tests antes de build, actions pineadas por SHA, permisos mínimos.
- `install.js`: tmp único, solo https, sin shell en `version`.

## v0.2.1

- Tests del camino feliz de `export` y lectura por pipe (guard inyectable).
- README con tabla comparativa vs alternativas.

## v0.2.0

- `hush export [NOMBRES]` — muestra valores solo en terminal real (por pipe se niega).
- Wizard `hush setup` — detecta claude/opencode/cursor/codex y propone registrar el MCP con confirmación y backup.
- `hush help [COMANDO]`, errores que nombran el problema, `status` con skills reales.
- Vault valida nombres y valores (rechaza multilínea y `=`).
- `install.js` verifica checksum también en el fallback de descarga.

## v0.1.1

- Fix: sin `bin` en package.json (el binario lo instala `install.js`).
- Docs: README en inglés, flag `--allow-scripts`.
- CI: publish npm por OIDC con fallback a token, `workflow_dispatch`.

## v0.1.0

- Vault (`~/.hush/vault`, `0600`), `set` por TTY/pipe, `check`/`list` solo nombres.
- `run`/`stdin` con inyección y salida redactada a `[REDACTED]`.
- MCP (`serve`): `hush_check`, `hush_list`, `hush_need` (elicitation), `hush_run`.
- `setup` instala el skill; `update` self-update verificado por checksum.
- Paquete npm `hush-secrets` con binarios por plataforma.
