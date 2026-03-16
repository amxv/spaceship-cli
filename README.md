# spaceship-cli

Simple Go CLI for common DNS operations on the Spaceship API.

## Endpoints Wrapped

- `GET /v1/domains` - domain list
- `GET /v1/domains/{domain}` - domain info
- `GET /v1/dns/records/{domain}` - domain DNS record list
- `PUT /v1/dns/records/{domain}` - save/update records
- `DELETE /v1/dns/records/{domain}` - delete records

## Required Spaceship API Scopes

- `domains:read`
- `dnsrecords:read`
- `dnsrecords:write`

Create an API key and secret in API Manager:
`https://www.spaceship.com/application/api-manager/`

## Install / Build

```bash
go build -o spaceship ./cmd/spaceship
```

## Authentication

Uses either:

1. Environment variables
- `SPACESHIP_API_KEY`
- `SPACESHIP_API_SECRET`

2. macOS Keychain (preferred)

```bash
./spaceship auth login
./spaceship auth status
./spaceship auth logout
```

Keychain service name: `spaceship-cli`

## Commands

### Domains

```bash
./spaceship domains list
./spaceship domains list --take 25 --skip 0 --order name
./spaceship domains info example.com
```

### DNS list

```bash
./spaceship dns list example.com
./spaceship dns list example.com --take 100 --skip 0 --order name
```

### DNS set (single-record helper)

```bash
# A record
./spaceship dns set example.com --type A --name @ --value 1.2.3.4 --ttl 300

# CNAME
./spaceship dns set example.com --type CNAME --name www --value target.example.com --ttl 300

# TXT
./spaceship dns set example.com --type TXT --name @ --value "v=spf1 a mx -all" --ttl 300

# MX
./spaceship dns set example.com --type MX --name @ --exchange mail.example.com --preference 10 --ttl 300
```

### DNS delete (single-record helper)

```bash
./spaceship dns delete example.com --type A --name @ --value 1.2.3.4
./spaceship dns delete example.com --type CNAME --name www --value target.example.com
```

### DNS put (bulk/raw)

Accepts either:
- A full payload object with `items` and optional `force`
- Or just an array of record items

```bash
./spaceship dns put example.com --file records.json
./spaceship dns put example.com --file records.json --force=true
```

## Notes

- API auth headers used:
  - `X-API-Key`
  - `X-API-Secret`
- Default API base URL: `https://spaceship.dev/api`
- API errors include Spaceship headers when available:
  - `spaceship-error-code`
  - `spaceship-operation-id`
