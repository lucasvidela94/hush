# hush — el agente usa secretos sin verlos

```bash
openssl rand -hex 24 | hush set WHATSAPP_VERIFY_TOKEN
hush check WHATSAPP_VERIFY_TOKEN META_APP_SECRET
hush stdin WHATSAPP_VERIFY_TOKEN -- npx wrangler secret put WHATSAPP_VERIFY_TOKEN
hush run --only META_APP_SECRET -- npx wrangler deploy
```

`check` y `list` solo muestran nombres. `run` y `stdin` inyectan en el
proceso hijo y redactan la salida a `[REDACTED]`.

## Desarrollo

```bash
make check   # fmt + vet + test + build
make hook    # instala pre-commit (fmt + vet + test)
```

Commits convencionales: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`.
Ver `SKILL.md` para el contrato del agente.
