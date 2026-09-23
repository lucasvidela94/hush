# Changelog

## v0.5.3

- `hush doctor`: chequeo de binario, vault, skills y registro MCP.
- Templates de issues/PRs, notas de plataforma en README.

## v0.5.2

- Redacción: wraps con padding, `token=`/`%` con runs aislados, base64 del
  reverso y hex del reverso, stride crudo 2-4, octal y decimal en paralelo,
  guard <5 contra falsos positivos.

## v0.5.1

- Fix: normalizados vacíos ya no anulan salidas (guard <4 real).
- Señales SIGTERM/SIGHUP/SIGQUIT también restauran el eco.
- Redacción: base32, bytes decimales, stride-2, join alfabético (od/rev con
  espacios ya no fugan).
- Docs: modelo shell, reglas de harness recomendadas, política de output.

## v0.5.0

- `hush_run` sin bypass: eliminado `confirm`, sin elicitation no ejecuta.
- Firma obligatoria en `update` (fail-closed) + lista de claves para rotar.
- Redacción por decodificación real (Basic auth, wraps, base64url, doble
  base64) y normalización total en el chequeo fail-closed.
- `hush set` restaura el eco ante Ctrl-C.
- Releases con environment `release` (aprobación) + attestations.

## v0.4.3

- Releases firmados automáticamente en CI (clave en secret `HUSH_SIGN_PRIV`).

## v0.4.2

- Nueva clave pública de firmas embebida.

## v0.4.1

- Releases firmados con ed25519 offline (`hush-sign`); `update` verifica
  `checksums.txt.sig` y solo acepta releases sin firma con aviso.

## v0.4.0

- `hush_run` pide confirmación humana por elicitation (comando + secretos).
- Redacción por decodificación (Basic auth y hex con saltos ya no fugan) y
  fail-closed siempre, no solo sin match literal.
- Runner mata el grupo de procesos completo y MCP corta a los 5 minutos.
- Vault versionado (`# hush-vault v2`, legacy intacto) + lock contra lost updates.
- Firmas ed25519 offline para releases (`hush-sign`, clave pública embebida).
- `x/term` para lectura sin eco (restore garantizado), multilínea con escapes.
- CI: attestations de provenance, sin token fallback.

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
