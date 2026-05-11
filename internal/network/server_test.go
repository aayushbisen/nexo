package network

import (
	"bufio"
	"context"
	"net"
	"nexo/internal/resp"
	"nexo/internal/store"
	"sync"
	"testing"
	"time"
)

// helper: creates a server with a fresh store for testing
func testServer(t *testing.T) *Server {
	return &Server{St: store.New[string](10)}
}

// helper: sends a command and reads one line of response
func sendAndRead(t *testing.T, conn net.Conn, cmd string) string {
	t.Helper()
	_, err := conn.Write([]byte(cmd + "\n"))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	sc := bufio.NewScanner(conn)
	if !sc.Scan() {
		t.Fatalf("no response for cmd %q", cmd)
	}
	return sc.Text()
}

// SET a key, then GET it back — verifies the full round trip.
func TestHandleConnectionSetThenGet(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "SET name nexo")
	if resp != "OK" {
		t.Fatalf("expected OK, got %q", resp)
	}

	resp = sendAndRead(t, client, "GET name")
	if resp != "nexo" {
		t.Fatalf("expected 'nexo', got %q", resp)
	}

	server.Close()
	wg.Wait()
}

// GET a key that doesn't exist returns an error.
func TestHandleConnectionGetMissing(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "GET missing")
	if resp != "Error key not found" {
		t.Fatalf("expected error, got %q", resp)
	}

	server.Close()
	wg.Wait()
}

// DEL removes a key from the store.
func TestHandleConnectionDelete(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	s.St.Set("name", "nexo")

	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "DEL name")
	if resp != "OK" {
		t.Fatalf("expected OK, got %q", resp)
	}

	_, ok := s.St.Get("name")
	if ok {
		t.Fatal("expected key to be deleted")
	}

	server.Close()
	wg.Wait()
}

// An unknown command returns an error.
func TestHandleConnectionUnknownCommand(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "INVALID")
	if resp != "Error Unknown command" {
		t.Fatalf("expected error, got %q", resp)
	}

	server.Close()
	wg.Wait()
}

// SET with missing arguments returns an error, and the connection stays open.
func TestHandleConnectionSetMissingArgs(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "SET")
	if resp != "Set command needs key and value to work" {
		t.Fatalf("expected error, got %q", resp)
	}

	server.Close()
	wg.Wait()
}

// GET with missing arguments returns an error.
func TestHandleConnectionGetMissingArgs(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "GET")
	if resp != "Get command needs key to work" {
		t.Fatalf("expected error, got %q", resp)
	}

	server.Close()
	wg.Wait()
}

// DEL with missing arguments returns an error.
func TestHandleConnectionDeleteMissingArgs(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "DEL")
	if resp != "Delete command needs key to work" {
		t.Fatalf("expected error, got %q", resp)
	}

	server.Close()
	wg.Wait()
}

// Cancelling the context causes Start to shut down the listener and return.
func TestStartShutsDownOnContextCancel(t *testing.T) {
	s := &Server{St: store.New[string](10), Port: 0}

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	done := make(chan struct{})
	go func() {
		s.Start(ctx, &wg)
		close(done)
	}()

	cancel()
	<-done
	wg.Wait()
}

// SET with EX stores a key that expires after the TTL.
func TestHandleConnectionSetWithEX(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "SET temp val EX 1")
	if resp != "OK" {
		t.Fatalf("expected OK, got %q", resp)
	}

	val, ok := s.St.Get("temp")
	if !ok {
		t.Fatal("expected key to exist immediately after SET")
	}
	if val != "val" {
		t.Fatalf("expected 'val', got %q", val)
	}

	time.Sleep(1100 * time.Millisecond)
	_, ok = s.St.Get("temp")
	if ok {
		t.Fatal("expected key to expire after TTL")
	}

	server.Close()
	wg.Wait()
}

// EXPIRE updates the TTL of an existing key.
func TestHandleConnectionExpire(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	s.St.Set("key", "value")

	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "EXPIRE key 1")
	if resp != "1" {
		t.Fatalf("expected '1', got %q", resp)
	}

	time.Sleep(1100 * time.Millisecond)
	_, ok := s.St.Get("key")
	if ok {
		t.Fatal("expected key to expire after EXPIRE")
	}

	server.Close()
	wg.Wait()
}

// EXPIRE on a missing key returns 0.
func TestHandleConnectionExpireMissing(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	resp := sendAndRead(t, client, "EXPIRE missing 10")
	if resp != "0" {
		t.Fatalf("expected '0', got %q", resp)
	}

	server.Close()
	wg.Wait()
}

// RESP SET with EX stores a key that expires after the TTL.
func TestHandleConnectionRESPSetWithEX(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	cmd := "*5\r\n$3\r\nSET\r\n$3\r\nkey\r\n$3\r\nval\r\n$2\r\nEX\r\n$1\r\n1\r\n"
	_, err := client.Write([]byte(cmd))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	val, err := resp.Read(bufio.NewReader(client))
	if err != nil {
		t.Fatalf("resp read failed: %v", err)
	}
	if val.Kind != '+' || val.Str != "OK" {
		t.Fatalf("expected +OK, got %c %q", val.Kind, val.Str)
	}

	time.Sleep(1100 * time.Millisecond)
	_, ok := s.St.Get("key")
	if ok {
		t.Fatal("expected key to expire after RESP SET EX")
	}

	server.Close()
	wg.Wait()
}

// RESP EXPIRE returns an integer reply.
func TestHandleConnectionRESPExpire(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := testServer(t)
	s.St.Set("key", "value")

	var wg sync.WaitGroup
	wg.Add(1)
	go s.handleConnection(server, &wg)

	cmd := "*3\r\n$6\r\nEXPIRE\r\n$3\r\nkey\r\n$2\r\n10\r\n"
	_, err := client.Write([]byte(cmd))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	val, err := resp.Read(bufio.NewReader(client))
	if err != nil {
		t.Fatalf("resp read failed: %v", err)
	}
	if val.Kind != ':' || val.Integer != 1 {
		t.Fatalf("expected :1, got %c %d", val.Kind, val.Integer)
	}

	server.Close()
	wg.Wait()
}
