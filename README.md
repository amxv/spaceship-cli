# spaceship-domains-cli

Simple CLI for Spaceship domains + DNS.

## Install

```bash
npm i -g spaceship-domains-cli
```

Then run:

```bash
spaceship --help
```

## Required API scopes

- `domains:read`
- `dnsrecords:read`
- `dnsrecords:write`

## Auth

```bash
spaceship auth login
spaceship auth status
spaceship auth logout
```

Credentials are stored in macOS Keychain (service: `spaceship-cli`).
You can also use env vars:

- `SPACESHIP_API_KEY`
- `SPACESHIP_API_SECRET`

## Common commands

```bash
# domains
spaceship domains list
spaceship domains info example.com

# dns
spaceship dns list example.com
spaceship dns set example.com --type A --name @ --value 1.2.3.4 --ttl 300
spaceship dns delete example.com --type A --name @ --value 1.2.3.4

# bulk update from file
spaceship dns put example.com --file records.json --force=true
```

## Release

Create tag `vX.Y.Z` and push. GitHub Actions will:

1. run quality checks
2. build cross-platform binaries and attach them to GitHub release
3. publish npm package `spaceship-domains-cli`

Set repository secret: `NPM_TOKEN`.
