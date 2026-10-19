---
description: "Use secrets without ever seeing them: check, need, run with redaction. Never accept pasted secrets."
---

# hush — el agente usa secretos sin verlos

## Regla cero: el chat no es un canal para secretos

Si el humano pega algo con pinta de secreto (token, key, hex largo, `sk-`, `ghp_`, `xoxb-`, app secret):

1. **No lo uses.** No lo pases a ningún comando, no lo re-emitas, no lo guardes en archivos.
2. Decile que ese valor ya quedó en el transcript y conviene **rotarlo** (regenerarlo en su dashboard).
3. Pedile que lo cargue por el canal seguro: en su propia terminal `hush set NOMBRE`, o vía `hush_need` para que lo escriba en el prompt nativo del harness.

Nunca digas "pasame el valor". Siempre: "corre `hush set NOMBRE` en tu terminal".

## Contrato

- Ninguna tool devuelve valores. `check`/`list` devuelven solo nombres.
- `need` guarda sin mostrar. `run` inyecta en el hijo y redacta a `[REDACTED]`.
- Los nombres también son contexto: `list` y `check` los muestran, y uno como `JIRATOKEN` ya dice qué sistema usa el humano. No los repitas fuera de lo necesario.
- Si falta un secreto en modo no interactivo: indicá `hush set NOMBRE`, no lo pidas por chat.

## Nunca leas de vuelta

La redacción solo cubre lo que pasa por `hush run`/`hush stdin`. Si un comando guarda el valor en un archivo (config de un MCP, `.env`, un remote de git con token), un comando posterior que lo imprima va directo al transcript.

- Después de cablear un secreto, verificá con un comando que no imprima la configuración (`claude mcp list`), o leelo a través de `hush run`.
- No ejecutes fuera de hush: `claude mcp get`, `config list`, `env`, `printenv`, `cat` de configs, `git remote -v` con token en la URL.
- Preferí referencias por variable de entorno (`${NOMBRE}`) en vez de valores persistidos, y lanzá la herramienta con `hush run --only NOMBRE -- ...`.
- Si `hush run` avisa "quedó guardado en texto plano en RUTA", el valor está en ese archivo: no lo leas, avisale al humano para que lo rote o lo cambie por una referencia.
- Si un valor llegó al transcript igual, avisale al humano que lo rote.

## Vía MCP (dentro del harness)

```
hush_check(names)              → {missing[], present[]}
hush_list()                    → nombres
hush_need(name, hint?)         → pide el valor por prompt nativo y lo guarda.
                                 hint: dónde lo encuentra el humano.
hush_run(command[], only[]?, all?, stdin_name?) → ejecuta con inyección + redacción.
                                 Deny-by-default: exige only[] o all=true. SIEMPRE pide
                                 confirmación humana por prompt antes de ejecutar.
                                 stdin_name: para comandos que leen por stdin
                                 (ej: ["npx","wrangler","secret","put","X"] con stdin_name X)
```

La confirmación humana es de la vía MCP. `hush run` por el shell solo pasa por el permiso normal del harness.

## Vía CLI (terminal del humano o tu shell)

```bash
hush check WHATSAPP_VERIFY_TOKEN META_APP_SECRET
hush run --only CLOUDFLARE_API_TOKEN -- npx wrangler secret list
hush stdin WHATSAPP_VERIFY_TOKEN -- npx wrangler secret put WHATSAPP_VERIFY_TOKEN
hush run --only META_APP_SECRET -- npx wrangler deploy
# Procesos longevos con stdio en streaming (servidores MCP): el modo normal
# retiene la salida hasta el exit; --stream la reenvía por línea con redacción.
hush run --stream --only JIRA_API_TOKEN -- jira-mcp
```

Humano en su terminal (no vos):
```bash
openssl rand -hex 24 | hush set WHATSAPP_VERIFY_TOKEN
hush set META_APP_SECRET
```

## Flujo WhatsApp (ejemplo)

1. `hush_check` → faltan los dos.
2. Humano: `hush set` x2 en su terminal (o `hush_need` con hint del dashboard).
3. `hush_run` con `stdin_name` para cada `wrangler secret put`, luego deploy.
4. Verificás el GET de Meta en logs. El transcript queda publicable.
