# spaceship-domains-cli

CLI for managing Spaceship domains and DNS records.

## 1. Install

```bash
npm i -g spaceship-domains-cli
```

This installs the `spaceship` command globally.

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

- Credentials are stored in macOS Keychain (service: `spaceship-cli`).
- You can also set credentials via env vars:
  - `SPACESHIP_API_KEY`
  - `SPACESHIP_API_SECRET`
- This CLI intentionally does not include commands for registering/deleting domains, nameserver management, transfers, SellerHub, or other broader account-context changes. It is intentionally scoped for safer agent-driven DNS/domain-read workflows.
