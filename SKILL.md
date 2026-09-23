# hush — skill para agentes

El agente **orquesta** secretos, nunca los **ve**.

## Reglas duras

1. Nunca pidas que te peguen un secreto en el chat.
2. Nunca hagas `cat .env* / .dev.vars`, `echo $SECRET`, `printenv | grep`, `wrangler secret put` con valor inline.
3. Si falta un secreto: corre `hush check`, y dile al humano exactamente: `hush set NOMBRE`.
4. Para usar: `hush run --` o `hush stdin --`. Nada más.

## Recetas

```bash
hush check WHATSAPP_VERIFY_TOKEN META_APP_SECRET
hush list
hush run -- npx wrangler secret list
hush stdin WHATSAPP_VERIFY_TOKEN -- npx wrangler secret put WHATSAPP_VERIFY_TOKEN
hush stdin META_APP_SECRET -- npx wrangler secret put META_APP_SECRET
hush run --only META_APP_SECRET -- npx wrangler deploy
```

Humano (en su terminal, no vos):
```bash
openssl rand -hex 24 | hush set WHATSAPP_VERIFY_TOKEN
hush set META_APP_SECRET
```

## Por qué

`hush set` solo lee de TTY o pipe del humano. `hush run/stdin` inyecta en el proceso hijo y redacta `[REDACTED]` antes de devolverte el output. El transcript queda publicable.
