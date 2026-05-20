# Nexo Examples

This folder contains runnable examples demonstrating how to connect to a running Nexo instance using various Redis client libraries.

## Prerequisites

1. **Start Nexo**: Run `make run` or `go run cmd/nexo/main.go` from the project root
2. Nexo listens on `localhost:9090` (coordinator) by default
3. The coordinator supports the Redis RESP protocol, so any standard Redis client works

## Supported Commands

| Command | Description |
|---------|-------------|
| `PING [msg]` | Returns PONG or the message |
| `SET key value [EX sec\|PX ms]` | Store a value with optional TTL |
| `GET key` | Retrieve a value |
| `DEL key` | Delete a key |
| `EXPIRE key seconds` | Set TTL on an existing key |

## Examples

### Go
```bash
cd go
go mod tidy
go run main.go
```

### Python
```bash
cd python
pip install -r requirements.txt
python main.py
```

### Node.js
```bash
cd node
npm install
node main.js
```

### Raw TCP (netcat)
```bash
# Plain text
echo "PING" | nc localhost 9090
echo "SET mykey hello" | nc localhost 9090
echo "GET mykey" | nc localhost 9090

# RESP protocol
printf '*1\r\n$4\r\nPING\r\n' | nc localhost 9090
printf '*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$5\r\nhello\r\n' | nc localhost 9090
printf '*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n' | nc localhost 9090
```

See `tcp/README.md` for more raw protocol examples.
