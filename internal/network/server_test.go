package network

import (
	"bufio"
	"context"
	"net"
	"nexo/internal/store"
	"sync"
	"testing"
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
