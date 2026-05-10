# Nexo

A lightweight, distributed, in-memory key-value cache written in Go. It features a type-safe generic core, LRU eviction, and consistent hashing across a TCP cluster.

## Architecture

```
Client → Coordinator (Hash Ring) → Worker Node(s)
                                      ↓
                              Generic LRU Store
```

- **`internal/store`** — Thread-safe generic LRU cache (`map` + `container/list`).
- **`internal/network`** — TCP server and coordinator proxy. Supports graceful shutdown via `signal.NotifyContext` and `sync.WaitGroup`.
- **`internal/hashring`** — CRC32-based consistent hashing ring with virtual nodes (`replicas` configurable per node) for even key distribution.
- **`internal/resp`** — RESP (REdis Serialization Protocol) Value type, Writer, and Reader.

## Quick Start

```bash
cd nexo
go run cmd/nexo/main.go
```

This starts:
- **Coordinator** on port `9090`
- **4 Workers** on ports `9091–9094` (each with 5 virtual nodes)

## Protocol

Two protocols supported on the same port. Detection is automatic via the first byte.

### RESP (REdis Serialization Protocol)

Use with `redis-cli` or raw `printf`:

```bash
# SET
printf '*3\r\n$3\r\nSET\r\n$1\r\na\r\n$1\r\n1\r\n' | nc localhost 9090
# +OK\r\n

# GET
printf '*2\r\n$3\r\nGET\r\n$1\r\na\r\n' | nc localhost 9090
# $1\r\n1\r\n
```

### Plain Text (nc-compatible)

Standard newline-delimited commands:

```
nc localhost 9090
SET name nexo
GET name
DEL name
```

## Structure

```
cmd/nexo/main.go          # Entrypoint (starts coordinator + workers)
internal/
  store/store.go          # Generic LRU store
  network/
    server.go             # Worker TCP server
    coordinator.go        # Routing proxy
  hashring/ring.go        # Consistent hash ring
  resp/resp.go            # RESP protocol type, reader, writer
```

## Make Commands

```bash
make build   # compile the binary to ./bin/nexo
make run     # build and run the cluster
make test    # run all tests
make format  # format all Go code
make clean   # remove build artifacts
```

## Running Tests

```bash
make test        # run all tests
make test -race  # check for race conditions
```

38 unit tests across all packages. **94.9% coverage.**

## Stack

- Go 1.26+
- `sync.RWMutex` for concurrency
- `container/list` for O(1) LRU tracking
- `net` + `bufio` for TCP networking

---

*Built for learning generics, networking, and distributed systems in Go.*

