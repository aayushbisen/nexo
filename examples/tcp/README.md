# Raw TCP/RESP Examples

Nexo auto-detects the protocol based on the first byte of the connection. If the first byte is `*`, it treats the connection as RESP. Otherwise, it falls back to plain text.

## Plain Text Protocol

Simple newline-delimited commands. No RESP framing required.

```bash
# PING
echo "PING" | nc localhost 9090
# Output: PONG

# PING with message
echo "PING hello" | nc localhost 9090
# Output: hello

# SET
echo "SET mykey myvalue" | nc localhost 9090
# Output: OK

# GET
echo "GET mykey" | nc localhost 9090
# Output: myvalue

# SET with EX (seconds)
echo "SET tempkey tempvalue EX 10" | nc localhost 9090
# Output: OK

# SET with PX (milliseconds)
echo "SET tempkey tempvalue PX 5000" | nc localhost 9090
# Output: OK

# EXPIRE
echo "EXPIRE mykey 30" | nc localhost 9090
# Output: 1

# DEL
echo "DEL mykey" | nc localhost 9090
# Output: OK

# GET after DEL (key not found)
echo "GET mykey" | nc localhost 9090
# Output: Error key not found

# Unknown command
echo "HGET mykey field" | nc localhost 9090
# Output: Error Unknown command
```

## RESP Protocol

Use `printf` to send properly formatted RESP messages.

```bash
# PING (no args)
printf '*1\r\n$4\r\nPING\r\n' | nc localhost 9090
# Output: +PONG\r\n

# PING with message
printf '*2\r\n$4\r\nPING\r\n$5\r\nhello\r\n' | nc localhost 9090
# Output: $5\r\nhello\r\n

# SET
printf '*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n' | nc localhost 9090
# Output: +OK\r\n

# GET
printf '*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n' | nc localhost 9090
# Output: $7\r\nmyvalue\r\n

# SET with EX
printf '*5\r\n$3\r\nSET\r\n$7\r\ntempkey\r\n$9\r\ntempvalue\r\n$2\r\nEX\r\n$2\r\n10\r\n' | nc localhost 9090
# Output: +OK\r\n

# EXPIRE
printf '*3\r\n$6\r\nEXPIRE\r\n$5\r\nmykey\r\n$2\r\n30\r\n' | nc localhost 9090
# Output: :1\r\n

# DEL
printf '*2\r\n$3\r\nDEL\r\n$5\r\nmykey\r\n' | nc localhost 9090
# Output: +OK\r\n

# GET after DEL (key not found)
printf '*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n' | nc localhost 9090
# Output: -key not found\r\n
```

## Interactive Session

For interactive testing, use `nc` without piping:

```bash
nc localhost 9090
# Type commands one at a time:
PING
SET foo bar
GET foo
DEL foo
GET foo
# Press Ctrl+C to exit
```

## RESP Format Reference

| Type | Prefix | Example | Meaning |
|------|--------|---------|---------|
| Simple String | `+` | `+OK\r\n` | Success response |
| Error | `-` | `-key not found\r\n` | Error response |
| Integer | `:` | `:1\r\n` | Numeric response |
| Bulk String | `$` | `$5\r\nhello\r\n` | String value |
| Array | `*` | `*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n` | Command with arguments |
