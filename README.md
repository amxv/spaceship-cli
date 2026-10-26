# spaceship-domains-cli

CLI for managing Spaceship domains and DNS records.

## 1. Install

```bash
npm i -g spaceship-domains-cli
```

This installs the `spaceship` command globally. Published npm packages include
prebuilt binaries for macOS (Apple Silicon and Intel), Linux (x64 and arm64),
and Windows (x64); a Go toolchain is not required for these platforms.

## 2. Create API key in Spaceship

Open API Manager:

- https://www.spaceship.com/application/api-manager/

Create an API key and enable these scopes:

- `domains:read`
- `dnsrecords:read`
- `dnsrecords:write`

## 3. Log in from CLI

```bash
spaceship auth login
```

Paste your API key and API secret when prompted.

Optional check:

```bash
spaceship auth status
```

## 4. Commands

### Domains

```bash
spaceship domains list
spaceship domains info example.com
```

### DNS

```bash
spaceship dns list example.com
spaceship dns set example.com --type A --name @ --value 1.2.3.4 --ttl 300
spaceship dns delete example.com --type A --name @ --value 1.2.3.4
spaceship dns put example.com --file records.json --force=true
```

## 5. Example: simple DNS change

Set apex A record to `1.2.3.4`:

```bash
spaceship dns set example.com --type A --name @ --value 1.2.3.4 --ttl 300
```

Then verify:

```bash
spaceship dns list example.com
```

## Notes

- On macOS, credentials are stored in Keychain (service: `spaceship-cli`).
- On Linux, credentials are stored in an owner-only file under the XDG config
  directory (normally `~/.config/spaceship-cli/credentials.json`). The
  directory uses permissions `0700`, and the file uses `0600`. This works on
  headless servers but is **not encrypted at rest**; disk encryption is
  recommended. The CLI refuses overly permissive paths and symbolic links.
- On Windows, use the environment variables below (interactive credential
  storage is not yet supported).
- You can also set credentials via env vars:
  - `SPACESHIP_API_KEY`
  - `SPACESHIP_API_SECRET`
  - When both are set, they take precedence over stored credentials.
- License: Apache-2.0.
- This CLI intentionally does not include commands for registering/deleting domains, nameserver management, transfers, SellerHub, or other broader account-context changes. It is intentionally scoped for safer agent-driven DNS/domain-read workflows.
