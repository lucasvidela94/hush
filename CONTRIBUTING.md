# Contributing

```bash
make check   # generate + fmt + vet + test + build (required before commit)
make hook    # installs the pre-commit hook
```

## Conventions

- Conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`, `ci:`.
- `cmd/` holds thin entrypoints; `internal/` holds decoupled packages
  (`vault`, `secret`, `redact`, `runner`, `cli`, `mcp`, `setup`, `update`).
- No unnecessary comments, no tautological names (`vault.Store`, not
  `vault.VaultStore`).
- Errors in Spanish for users, identifiers in English.
- `SKILL.md` is the single source of the agent skill; `go generate ./...`
  refreshes the embedded copy (a test enforces it).
- Release: `./scripts/release.sh X.Y.Z`, push branch + tag. The workflow
  builds, publishes npm (OIDC), and creates the GitHub release.

## Security-sensitive areas

Changes to `redact`, `runner`, `secret`, `vault`, `mcp/handlers.go`,
`install.js`, or `update` need a test proving no value leaks and a threat
note in the commit message or `SECURITY.md`.
