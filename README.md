# DNS server

dns-server is a simple DNS server created for using in local network.

#### dns-server can:

- Contain custom domains

#### To Do

- Refactor everything
- Change postgersql to sqlite
- Add normal CLI
- Add functionality to request domains to other DNS servers defined by user
- Add functionality to block domains

## Installation

With docker compose:
```
git clone https://github.com/s-588/dns-server
cd dns-server
docker compose up
```

Without docker:
```
go install https://github.com/s-588/dns-server
```

## Usage

Set your DNS to 127.0.0.1 in system settings or settings of your network.

Every request now should go through dns-server.

TODO
CLI commands


