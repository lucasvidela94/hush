# Privacy

- Secrets live only on your machine, in `~/.hush/vault` (file `0600`).
- `hush` has no telemetry, no accounts, no network calls — except:
  - `hush update`, which queries api.github.com and downloads from
    github.com releases (checksum-verified).
  - `install.js` fallback, which downloads the binary from GitHub releases
    when the npm tarball has no bundled asset (checksum-verified).
- Uninstall keeps the vault so reinstalling loses nothing. Wipe it with:
  `rm -rf ~/.hush`.
- MCP elicitation sends the typed value to the local `hush serve` process
  through your harness app. It never goes to the model in compliant clients,
  but prefer `hush set` in your own terminal for maximum privacy.
