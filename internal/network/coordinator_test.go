package network

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"nexo/internal/hashring"
	"sync"
	"testing"
)

// helper: creates a coordinator with a ring and injectable dial
func testCoordinator(t *testing.T, workerConn net.Conn) *Coordinator {
	t.Helper()
	hr := hashring.New()
	hr.AddNode("localhost:9091", 5)

	return &Coordinator{
		Port: 0,
		Ring: hr,
		Dial: func(_, _ string) (net.Conn, error) {
			return workerConn, nil
		},
	}
}

// helper: sends a command and reads one line of response
func coordSendAndRead(t *testing.T, conn net.Conn, cmd string) string {
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

// SET command is forwarded to the worker, and the worker's response is relayed back.
func TestCoordinatorRelaysSetResponse(t *testing.T) {
	client, server := net.Pipe()
	workerClient, workerServer := net.Pipe()

	c := testCoordinator(t, workerClient)

	var wg sync.WaitGroup
	wg.Add(1)
	go c.handleConnection(server, &wg)

	// Simulate worker response
	go func() {
		buf := bufio.NewReader(workerServer)
		cmd, _ := buf.ReadString('\n')
		if cmd != "SET key val\n" {
			t.Errorf("worker got unexpected command: %q", cmd)
		}
		workerServer.Write([]byte("OK\n"))
		workerServer.Close()
	}()

	resp := coordSendAndRead(t, client, "SET key val")
	if resp != "OK" {
		t.Fatalf("expected OK, got %q", resp)
	}

	client.Close()
	server.Close()
	wg.Wait()
}

// GET command is forwarded, worker response is relayed back.
func TestCoordinatorRelaysGetResponse(t *testing.T) {
	client, server := net.Pipe()
	workerClient, workerServer := net.Pipe()

	c := testCoordinator(t, workerClient)

	var wg sync.WaitGroup
	wg.Add(1)
	go c.handleConnection(server, &wg)

	go func() {
		buf := bufio.NewReader(workerServer)
		cmd, _ := buf.ReadString('\n')
		if cmd != "GET key\n" {
			t.Errorf("worker got unexpected command: %q", cmd)
		}
		workerServer.Write([]byte("value\n"))
		workerServer.Close()
	}()

	resp := coordSendAndRead(t, client, "GET key")
	if resp != "value" {
		t.Fatalf("expected 'value', got %q", resp)
	}

	client.Close()
	server.Close()
	wg.Wait()
}

// When the worker is unreachable, the coordinator returns an error.
func TestCoordinatorWorkerUnavailable(t *testing.T) {
	client, server := net.Pipe()

	hr := hashring.New()
	hr.AddNode("localhost:9999", 1)

	c := &Coordinator{
		Ring: hr,
		Dial: func(_, addr string) (net.Conn, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go c.handleConnection(server, &wg)

	resp := coordSendAndRead(t, client, "SET key val")
	if resp != "Error: Worker unavailable" {
		t.Fatalf("expected worker unavailable error, got %q", resp)
	}

	client.Close()
	server.Close()
	wg.Wait()
}

// SET with missing arguments returns an error before reaching the worker.
func TestCoordinatorSetMissingArgs(t *testing.T) {
	client, server := net.Pipe()

	hr := hashring.New()
	hr.AddNode("localhost:9091", 1)

	c := &Coordinator{Ring: hr, Dial: func(_, _ string) (net.Conn, error) {
		t.Fatal("dial should not be called for a bad SET")
		return nil, nil
	}}

	var wg sync.WaitGroup
	wg.Add(1)
	go c.handleConnection(server, &wg)

	resp := coordSendAndRead(t, client, "SET")
	if resp != "Set command needs key and value to work" {
		t.Fatalf("expected error, got %q", resp)
	}

	client.Close()
	server.Close()
	wg.Wait()
}

// DEL command is forwarded to the worker, response is relayed back.
func TestCoordinatorRelaysDelResponse(t *testing.T) {
	client, server := net.Pipe()
	workerClient, workerServer := net.Pipe()

	c := testCoordinator(t, workerClient)

	var wg sync.WaitGroup
	wg.Add(1)
	go c.handleConnection(server, &wg)

	go func() {
		buf := bufio.NewReader(workerServer)
		cmd, _ := buf.ReadString('\n')
		if cmd != "DEL key\n" {
			t.Errorf("worker got unexpected command: %q", cmd)
		}
		workerServer.Write([]byte("OK\n"))
		workerServer.Close()
	}()

	resp := coordSendAndRead(t, client, "DEL key")
	if resp != "OK" {
		t.Fatalf("expected OK, got %q", resp)
	}

	client.Close()
	server.Close()
	wg.Wait()
}

// GET with missing arguments returns an error before reaching the worker.
func TestCoordinatorGetMissingArgs(t *testing.T) {
	client, server := net.Pipe()

	hr := hashring.New()
	hr.AddNode("localhost:9091", 1)

	c := &Coordinator{Ring: hr, Dial: func(_, _ string) (net.Conn, error) {
		t.Fatal("dial should not be called for a bad GET")
		return nil, nil
	}}

	var wg sync.WaitGroup
	wg.Add(1)
	go c.handleConnection(server, &wg)

	resp := coordSendAndRead(t, client, "GET")
	if resp != "Get command needs key to work" {
		t.Fatalf("expected error, got %q", resp)
	}

	client.Close()
	server.Close()
	wg.Wait()
}

// DEL with missing arguments returns an error before reaching the worker.
func TestCoordinatorDeleteMissingArgs(t *testing.T) {
	client, server := net.Pipe()

	hr := hashring.New()
	hr.AddNode("localhost:9091", 1)

	c := &Coordinator{Ring: hr, Dial: func(_, _ string) (net.Conn, error) {
		t.Fatal("dial should not be called for a bad DEL")
		return nil, nil
	}}

	var wg sync.WaitGroup
	wg.Add(1)
	go c.handleConnection(server, &wg)

	resp := coordSendAndRead(t, client, "DEL")
	if resp != "Delete command needs key to work" {
		t.Fatalf("expected error, got %q", resp)
	}

	client.Close()
	server.Close()
	wg.Wait()
}

// When the worker connection succeeds but the response read fails, an error is returned.
func TestCoordinatorWorkerResponseFailed(t *testing.T) {
	client, server := net.Pipe()
	workerClient, workerServer := net.Pipe()

	c := testCoordinator(t, workerClient)

	// Close the server end so the coordinator's ReadString gets an error
	workerServer.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go c.handleConnection(server, &wg)

	resp := coordSendAndRead(t, client, "SET key val")
	if resp != "Error: Worker response failed" {
		t.Fatalf("expected worker response failed error, got %q", resp)
	}

	client.Close()
	server.Close()
	wg.Wait()
}

// An unknown command returns an error.
func TestCoordinatorUnknownCommand(t *testing.T) {
	client, server := net.Pipe()

	hr := hashring.New()
	hr.AddNode("localhost:9091", 1)

	c := &Coordinator{Ring: hr, Dial: func(_, _ string) (net.Conn, error) {
		t.Fatal("dial should not be called for unknown command")
		return nil, nil
	}}

	var wg sync.WaitGroup
	wg.Add(1)
	go c.handleConnection(server, &wg)

	resp := coordSendAndRead(t, client, "INVALID")
	if resp != "Error Unknown command" {
		t.Fatalf("expected error, got %q", resp)
	}

	client.Close()
	server.Close()
	wg.Wait()
}

// Cancelling the context causes Start to shut down the listener and return.
func TestCoordinatorStartShutsDownOnContextCancel(t *testing.T) {
	hr := hashring.New()
	hr.AddNode("localhost:9091", 1)

	c := &Coordinator{Port: 0, Ring: hr}
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	done := make(chan struct{})
	go func() {
		c.Start(ctx, &wg)
		close(done)
	}()

	cancel()
	<-done
	wg.Wait()
}
