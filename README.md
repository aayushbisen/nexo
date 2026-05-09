# Nexo

A lightweight, distributed, in-memory key-value cache written in Go. It features a type-safe generic core, LRU eviction, and consistent hashing across a TCP cluster.

## Architecture

```
Client → Coordinator (Hash Ring) → Worker Node(s)
                                      ↓
                              Generic LRU Store
```

- **`internal/store`** — Thread-safe generic LRU cache (`map` + `container/list`).
- **`internal/network`** — TCP server and coordinator proxy.
- **`internal/hashring`** — CRC32-based consistent hashing ring for key routing.

## Quick Start

```bash
cd nexo
go run cmd/nexo/main.go
```

This starts:
- **Coordinator** on port `9090`
- **4 Workers** on ports `9091–9094`

## Protocol

Connect via `nc` or `telnet`:

```
nc localhost 9090
SET name nexo
GET name
DEL name
```

Responses are newline-delimited (`OK`, value, or error).

## Structure

```
cmd/nexo/main.go          # Entrypoint (starts coordinator + workers)
internal/
  store/store.go          # Generic LRU store
  network/
    server.go             # Worker TCP server
    coordinator.go        # Routing proxy
  hashring/ring.go        # Consistent hash ring
```

## Stack

- Go 1.26+
- `sync.RWMutex` for concurrency
- `container/list` for O(1) LRU tracking
- `net` + `bufio` for TCP networking

---

*Built for learning generics, networking, and distributed systems in Go.*

