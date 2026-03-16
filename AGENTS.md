# AGENTS.md

This file is the operator guide for future coding agents working in this repository.

## Project Purpose

`spaceship-domains-cli` is a Go CLI that wraps key Spaceship API operations for domains and DNS management.

Current command surface:

- `spaceship auth <login|status|logout>`
- `spaceship domains <list|info>`
- `spaceship dns <list|set|delete|put>`
- `spaceship help ...` (built-in extensive help)

## API Scope (What this CLI uses)

Endpoints wrapped:

- `GET /v1/domains`
- `GET /v1/domains/{domain}`
- `GET /v1/dns/records/{domain}`
- `PUT /v1/dns/records/{domain}`
- `DELETE /v1/dns/records/{domain}`

Required Spaceship API scopes:

- `domains:read`
- `dnsrecords:read`
- `dnsrecords:write`

## Repo Layout

- `cmd/spaceship/main.go` - entrypoint
- `internal/app/` - CLI routing, flags, help output
- `internal/client/` - HTTP client for Spaceship API
- `internal/credentials/` - credential loading and keychain integration
- `internal/output/` - table/JSON output formatters
- `bin/spaceship.js` - npm global command shim (`spaceship`)
- `scripts/postinstall.js` - npm postinstall binary downloader/build fallback
- `.github/workflows/release.yml` - tag-driven release pipeline (GitHub release + npm publish)
- `Makefile` - local command runner for checks/build/release

## Local Toolchain Expectations

- Go: Homebrew Go (`go1.26+` recommended)
- Node.js: `>=18` (for npm wrapper scripts)
- Optional but recommended:
  - `golangci-lint`
  - `staticcheck`

## Primary Dev Commands

Use the `Makefile` as the canonical command runner.

- `make help` - list available tasks
- `make check` - full quality pass (`fmt`, `typecheck`, `test`, `vet`, `staticcheck`, `golangci-lint`, node lint)
- `make build` - build Go CLI to `dist/spaceship`
- `make install-local` - install local binary to `~/.local/bin/spaceship`
- `make smoke` - build and run `--help`

Direct commands (equivalent):

- `go build ./...` (typecheck)
- `go test ./...`
- `go vet ./...`
- `staticcheck ./...`
- `golangci-lint run`
- `npm run lint`

## Release Process

Releases are tag-driven.

1. Ensure local checks are green:
- `make check`

2. Push a release tag:
- `make release-tag TAG=vX.Y.Z`

3. CI workflow (`.github/workflows/release.yml`) will:
- run quality checks
- build cross-platform binaries and attach them to GitHub release
- publish npm package `spaceship-domains-cli`

Required GitHub secret:

- `NPM_TOKEN`

## Credentials and Auth Behavior

Credential lookup order:

1. `SPACESHIP_API_KEY` + `SPACESHIP_API_SECRET` env vars
2. macOS Keychain entries (service `spaceship-cli`)

Useful commands:

- `spaceship auth login`
- `spaceship auth status`
- `spaceship auth logout`

## Agent Workflow Guidance

When making changes:

1. Prefer updating built-in CLI help text alongside behavior changes.
2. Run `make check` before finalizing.
3. Rebuild and smoke test with `make smoke` for CLI-facing changes.
4. If command behavior changed, update `README.md` and this `AGENTS.md`.
5. Keep syntax simple and script-friendly (`--json` remains available for automation).

## Common Pitfalls

- Do not assume credentials exist; check with `spaceship auth status`.
- For DNS delete operations, record matching fields must align with API expectations.
- Keep npm and Go release versions in sync via tag release workflow (`vX.Y.Z`).
