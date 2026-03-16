# CONTRIBUTORS.md

Contributor/developer notes for `spaceship-domains-cli`.

## Local quality commands

Use the command runner:

```bash
make help
make check
```

Or run commands directly:

```bash
go build ./...
go test ./...
go vet ./...
staticcheck ./...
golangci-lint run
npm run lint
```

## Build and local install

```bash
make build
make install-local
make smoke
```

## Release flow

Release is tag-driven via GitHub Actions (`.github/workflows/release.yml`).

1. Ensure checks are green:

```bash
make check
```

2. Create and push tag:

```bash
make release-tag TAG=vX.Y.Z
```

3. CI does:
- quality checks
- cross-platform binary build + GitHub release assets
- npm publish (`spaceship-domains-cli`)

## Required secrets

GitHub repo secret:

- `NPM_TOKEN`

## Auth scopes for Spaceship API key

- `domains:read`
- `dnsrecords:read`
- `dnsrecords:write`
