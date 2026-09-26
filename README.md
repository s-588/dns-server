
# DNS Server

A Go DNS server that answers queries from PostgreSQL-stored records, exposes a protobuf HTTP CRUD API, and includes CLI + TUI administration tools.

## Key Features

- DNS server (UDP + TCP) backed by PostgreSQL
- HTTP API with protobuf serialization for records and users
- JWT cookie authentication + role-based access (`admin` / `user`)
- Prometheus metrics on `/metrics`
- Structured JSON logging + real-time WebSocket log streaming
- CLI for record management and a Bubble Tea TUI dashboard

## Built With

- Go 1.25+
- chi, gorilla/websocket, golang-jwt/jwt
- pgx, goose, sqlc
- Prometheus client
- Bubble Tea (TUI)

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Docker & Docker Compose (optional)

### Quick Start with Docker Compose

```bash
git clone https://github.com/prionis/dns-server
cd dns-server
docker compose up
```

- DNS: `localhost:53` (UDP/TCP)
- HTTP API: `localhost:8083`

> Port 53 must be free.

### Local Build

```bash
go build -o dns-server .
```

### Configuration

Create `dns-server.env` and `postgres.env` (or use the provided examples):

```
JWT_SECRET=your-secret-key
POSTGRES_USER=dnsuser
POSTGRES_PASSWORD=dnsuserpass
POSTGRES_DB=dnsdb
POSTGRES_ADDR=localhost   # or "postgres" inside Docker Compose
```

## Usage

The same binary acts as server or client.

### Start the server

```bash
./dns-server -server
```

Migrations run automatically. An initial admin user (`admin` / `admin`) is created if none exists.

### CLI

```bash
./dns-server -add "example.com. 3600 IN A 192.168.1.1"
./dns-server -del 42
./dns-server -records
./dns-server -logs
./dns-server -tui          # interactive dashboard (requires running server)
```

### Server flags

| Flag       | Description              | Default        |
|------------|--------------------------|----------------|
| `-addr`    | HTTP listen address      | `127.0.0.1`    |
| `-port`    | HTTP port                | `:8080`        |
| `-logfile` | Log file path            | `DNSServer.log`|

## HTTP API

Most endpoints require a JWT cookie obtained from `/auth/login`.

- **Auth**: `POST /auth/login`, `POST /auth/register` (admin)
- **Users** (admin): `GET/DELETE/PATCH /api/users…`
- **Resource Records**: full CRUD under `/api/rrs`
- **Logs**: `GET /api/logs/all`, WebSocket at `/api/logs/ws`
- **Metrics**: `GET /metrics`

Protobuf definitions live in `crud.proto`.
