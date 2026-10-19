# Changelog

## v0.8.0

- `hush setup` escribe `~/.codex/rules/hush.rules` para que Codex pida
  confirmación en `hush run`, `stdin` y `export`, aunque antes se haya
  aprobado un `allow`. `hush doctor` recomienda además el perfil de permisos
  que niega `~/.hush` en el sandbox de Codex (bloquea hasta `cat` por shell;
  hush se usa entonces por su servidor MCP).
- `hush setup` registra el servidor MCP en opencode v2 dentro de
  `mcp.servers.hush` (antes escribía `mcp.hush`, el formato v1, que v2 no
  lee) y mueve una entrada vieja si la encuentra. La versión sale de la
  config (`permission` → v1; `permissions` o `mcp.servers` → v2) o, sin
  marcas, de `opencode --version`. `hush doctor` chequea la ruta correcta y
  avisa si la entrada quedó en formato v1.
- `hush setup` ofrece agregar reglas de permisos para que el agente no lea ni
  edite `~/.hush` con sus herramientas y pida confirmación en `hush run`,
  `stdin` y `export`: Claude Code (`~/.claude/settings.json`) y opencode v2
  (`permissions`). Fusiona con las reglas existentes y deja backup.
  `hush doctor` avisa si faltan. Antes eran solo una recomendación en
  `SECURITY.md`, y sin ellas un agente con shell leía el vault con un `cat`.

## v0.7.0

- `hush run --stream --only A,B -- cmd` reenvía stdout/stderr por línea a
  medida que llegan, para procesos longevos con stdio en streaming (servidores
  MCP). Sin `--stream`, `hush run` sigue esperando al exit y por eso no servía
  como wrapper de un servidor MCP. Cada línea se tapa mirando también las
  anteriores: un secreto repartido entre líneas se corta antes de que salga
  más que un fragmento corto.
- La redacción detecta tramos de 12 caracteres de un secreto aunque estén
  separados por espacios o saltos de línea. Antes, de un token largo podía
  quedar visible hasta el 40% (por ejemplo `fold -w1 | head -40`) porque solo
  se buscaba el valor completo o una ventana proporcional a su largo.

## v0.6.1

- `hush update` ahora delega en npm: consulta la última versión y corre
  `npm install -g hush-secrets@VERSION`. Necesita npm en el PATH. Se eliminan
  las claves de firma embebidas, `hush-sign` y `checksums.txt.sig`: ya no hay
  claves que rotar ni que puedan dejar a los usuarios sin actualizar.
- **Reinstalá una vez** si tenés ≤v0.6.0: su `hush update` verifica una firma
  que los releases nuevos ya no publican y va a pedir
  `npm i -g hush-secrets@latest`. De v0.6.1 en adelante `hush update` funciona
  solo.

## v0.6.0

Cambios de comportamiento (revisá tus scripts):

- `hush run --only A,B` sale con código 3 y no ejecuta el comando si falta
  algún secreto (antes corría con la variable sin definir).
- `hush set` rechaza nombres que no sean `[A-Za-z_][A-Za-z0-9_]*`. Los
  secretos ya guardados con otros nombres siguen funcionando.
- Cuando se detecta un secreto transformado, solo se tapan las líneas
  afectadas y el marcador termina en newline (antes se reemplazaba toda la
  salida, sin newline).

Nuevo:

- `hush rm NOMBRE...` (alias `delete`, `unset`, `remove`) con confirmación en
  terminal y `--yes`. Si falta algún nombre no borra nada.
- `hush set` avisa si el valor tiene menos de 4 caracteres (no se tapa).
- `hush run` avisa, con nombre y ruta pero sin el valor, si el comando guardó
  un secreto en texto plano en `~/.claude.json`, `.env*`, `.mcp.json`,
  `opencode.json` o la config de cursor/opencode.
- Redacción de substrings (6+ caracteres y al menos 40% del valor),
  permutaciones de caracteres y copias enmascaradas. Puede dar falsos
  positivos con secretos cortos.
- `hush doctor` ya no rompe el layout con los snippets multilínea.
- `SKILL.md` y `SECURITY.md`: guía de "nunca leas de vuelta" y nuevos límites.

Rendimiento:

- `redact.Apply` deja de ser cuadrático: 1 MB de salida pasó de ~18 s a
  menos de 0,1 s.

Interno: CI en cada PR (vet, golangci-lint estricto, `test -race`).

## v0.5.6

- Rotación de clave de firma: se agrega segunda pubkey a `releasePubKeys`.
  Los binarios ≤v0.5.5 rechazan la firma nueva (fail-closed) y piden
  reinstalar por npm una vez (`npm i -g hush-secrets@latest`). v0.5.5 quedó solo en npm (sin GitHub Release).

## v0.5.5

- `hush set` muestra `*` por caracter (Backspace borra uno, Ctrl-U limpia
  la línea). Evita pegar dos veces sin darse cuenta; el valor sigue sin
  salir en el transcript.

## v0.5.4

- README con matriz de harnesses verificada en vivo.

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
