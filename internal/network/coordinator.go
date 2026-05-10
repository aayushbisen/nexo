package network

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"nexo/internal/hashring"
	"nexo/internal/resp"
	"strings"
	"sync"
)

type Coordinator struct {
	Port int
	Ring *hashring.Ring
	Dial func(network, addr string) (net.Conn, error)
}

func New() *Coordinator {
	return &Coordinator{}
}

func (c *Coordinator) handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer conn.Close()
	defer wg.Done()
	r := bufio.NewReader(conn)

	for {
		first, err := r.Peek(1)
		if err != nil {
			break
		}
		var cmd string
		var args []string
		respMode := first[0] == '*'

		if respMode {
			val, err := resp.Read(r)
			if err != nil {
				break
			}
			if len(val.Array) < 1 {
				continue
			}
			cmd = val.Array[0].Str
			for _, a := range val.Array[1:] {
				args = append(args, a.Str)
			}
		} else {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			parts := strings.Fields(strings.TrimRight(line, "\r\n"))
			if len(parts) == 0 {
				continue
			}
			cmd = parts[0]
			args = parts[1:]
		}

		switch cmd {
		case "SET":
			if len(args) < 2 {
				writeClientError(conn, "Set command needs key and value to work", respMode)
				continue
			}
		case "GET":
			if len(args) < 1 {
				writeClientError(conn, "Get command needs key to work", respMode)
				continue
			}
		case "DEL":
			if len(args) < 1 {
				writeClientError(conn, "Delete command needs key to work", respMode)
				continue
			}
		default:
			writeClientError(conn, "Error Unknown command", respMode)
			continue
		}

		addr := c.Ring.GetNode(args[0])
		workerConn, err := c.Dial("tcp", addr)
		if err != nil {
			writeClientError(conn, "Error: Worker unavailable", respMode)
			continue
		}

		wargs := make([]resp.Value, 0, 1+len(args))
		wargs = append(wargs, resp.Value{Kind: '$', Str: cmd})
		for _, a := range args {
			wargs = append(wargs, resp.Value{Kind: '$', Str: a})
		}
		resp.Value{Kind: '*', Array: wargs}.Write(workerConn)

		response, err := resp.Read(bufio.NewReader(workerConn))
		workerConn.Close()
		if err != nil {
			writeClientError(conn, "Error: Worker response failed", respMode)
			continue
		}

		if respMode {
			response.Write(conn)
		} else {
			switch response.Kind {
			case '+':
				fmt.Fprintf(conn, "%s\n", response.Str)
			case '$':
				fmt.Fprintf(conn, "%s\n", response.Str)
			case '-':
				fmt.Fprintf(conn, "Error: %s\n", response.Str)
			default:
				io.WriteString(conn, "Error: unexpected response\n")
			}
		}
	}
}

func writeClientError(conn net.Conn, msg string, respMode bool) {
	if respMode {
		resp.Value{Kind: '-', Str: msg}.Write(conn)
	} else {
		io.WriteString(conn, msg+"\n")
	}
}

func (c *Coordinator) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	if c.Dial == nil {
		c.Dial = net.Dial
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		return
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()

		if err != nil {
			return
		}
		wg.Add(1)
		go c.handleConnection(conn, wg)
	}

}
