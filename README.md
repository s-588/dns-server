# DNS Server

A DNS server implementation in Go that handles queries, exposes a CRUD API over HTTP, and provides a CLI + TUI for administration.  
It uses PostgreSQL for persistent storage, supports JWT authentication and role-based access, and can stream logs via WebSockets.

## What it does

- Responds to DNS queries (UDP and TCP) for records stored in PostgreSQL.
- Provides an HTTP API to manage resource records and users with protobuf serialization.
- Authenticates users via JWT cookies; roles (`admin`, `user`) control access.
- Exports Prometheus metrics on `/metrics`.
- Logs in JSON format to file and stdout; real‑time log streaming via WebSocket.
- Offers a CLI for common operations (add/delete records, view logs/records) and a TUI for interactive monitoring.

## Built with

- Go 1.25.7
- [chi](https://github.com/go-chi/chi) – HTTP router
- [gorilla/websocket](https://github.com/gorilla/websocket) – WebSockets
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) – JWT handling
- [pgx](https://github.com/jackc/pgx) – PostgreSQL driver
- [goose](https://github.com/pressly/goose) – migrations
- [Prometheus client](https://github.com/prometheus/client_golang) – metrics
- [sqlc](https://sqlc.dev/) – type‑safe SQL generation
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) – TUI

## Getting started

### Prerequisites

- Go 1.25.7+
- PostgreSQL 15+
- (Optional) Docker and Docker Compose

### Clone & build

```bash
git clone https://github.com/prionis/dns-server
cd dns-server
go build -o dns-server .
```

### Configuration

Copy the example environment files (or create your own):

**dns-server.env**
```
JWT_SECRET=your-secret-key
```

**postgres.env**
```
POSTGRES_USER=dnsuser
POSTGRES_PASSWORD=dnsuserpass
POSTGRES_DB=dnsdb
POSTGRES_ADDR=postgres   # or localhost if run outside of Docker Compose
```

When running with Docker Compose, adjust `POSTGRES_ADDR` to match the service name (`postgres`).  
For local development, set `POSTGRES_ADDR=localhost` and ensure the database is reachable on port 5432.

### Running with Docker Compose

```bash
docker compose up
```

This starts the server on:
- DNS: `localhost:53` (UDP/TCP)
- HTTP API: `localhost:8083`

Server won't start if any app already uses port 53.

## Usage

The same binary can act as the server or as a client.

### Server modes

Start the DNS + HTTP server:

```bash
./dns-server -server
```

The server automatically runs database migrations and creates an initial `admin` user (`login: admin, password: admin`) if none exists.

### CLI client

Add a resource record (RFC 1035 format):

```bash
./dns-server -add "example.com. 3600 IN A 192.168.1.1"
```

Delete a record by its database ID:

```bash
./dns-server -del 42
```

List all records:

```bash
./dns-server -records
```

View logs:

```bash
./dns-server -logs
```

Start the TUI (interactive dashboard):

```bash
./dns-server -tui
```

The TUI requires a running server – it connects to the HTTP API and WebSocket endpoint.  
You will be prompted to log in, use `admin` `admin` as login and password.

### Server flags

| Flag | Description | Default |
|------|-------------|---------|
| `-addr` | HTTP server address | `127.0.0.1` |
| `-port` | HTTP server port | `:8080` |
| `-logfile` | Path to log file | `DNSServer.log` |

## HTTP API

All endpoints (except `/auth/*` and `/metrics`) require a valid JWT cookie set by login.

### Authentication

- `POST /auth/login` – expects a protobuf `Login` message; returns a `User` and sets a cookie.
- `POST /auth/register` – admin only; expects `Register`; creates a user.

### Users (admin only)

- `GET /api/users/all` – list all users
- `GET /api/users/{id}` – get one user
- `DELETE /api/users` – delete user (ID in request body)
- `PATCH /api/users` – update user (protobuf `User`)

### Resource records (authenticated)

- `GET /api/rrs/all` – list all records
- `GET /api/rrs/{id}` – get one record
- `POST /api/rrs` – create record (protobuf `ResourceRecord`)
- `PATCH /api/rrs/{id}` – update record (protobuf `ResourceRecord`)
- `DELETE /api/rrs/{id}` – delete record

### Logs (authenticated)

- `GET /api/logs/all` – returns all log entries as protobuf `LogCollection`
- `GET /api/logs/ws` – WebSocket endpoint; each message is a JSON‑formatted log line

### Metrics

- `GET /metrics` – Prometheus metrics

All protobuf definitions are in `crud.proto`; generated Go code is in `proto/crud/genproto/`.

## Metrics exposed

- `http_requests_total` – method, path, status
- `http_request_duration_seconds` – method, path
- `dns_queries_total` – query type, response code
- `dns_query_duration_seconds` – query type
- `dns_records_found_total` – query type
- `login_attempts_total` – result (`success`, `invalid_user`, `wrong_password`)
- `rr_operations_total` – operation (`get`, `post`, `patch`, `delete`) and result (`success`, `error`)

## Database schema

Tables: `types`, `classes`, `resource_records`, `roles`, `users`.  
Migrations are embedded and applied at startup.

The `resource_records` table stores domain, data, type/class foreign keys, and TTL.  
Users store bcrypt‑hashed passwords and a role reference.
