package network

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"nexo/internal/hashring"
	"strings"
	"sync"
)

type Coordinator struct {
	Port int
	Ring *hashring.Ring
}

func New() *Coordinator {
	return &Coordinator{}
}

func (c *Coordinator) handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer conn.Close()
	defer wg.Done()
	sc := bufio.NewScanner(conn)

	for sc.Scan() {
		cmd := sc.Text()
		listCmd := strings.Fields(cmd)
		key := ""
		switch listCmd[0] {
		case "SET":
			if len(listCmd) < 3 {
				io.WriteString(conn, "Set command needs key and value to work\n")
				continue
			}
			key = listCmd[1]

			// s.St.Set(listCmd[1], listCmd[2])
			// io.WriteString(conn, "OK\n")
		case "GET":
			if len(listCmd) < 2 {
				io.WriteString(conn, "Get command needs key to work\n")
				continue
			}
			key = listCmd[1]

			// val, ok := s.St.Get(listCmd[1])
			// if ok {
			// 	fmt.Fprintf(conn, "%s\n", val)
			// } else {
			// 	io.WriteString(conn, "Error key not found\n")
			// }
		case "DEL":
			if len(listCmd) < 2 {
				io.WriteString(conn, "Delete command needs key to work\n")
				continue
			}
			key = listCmd[1]

			// s.St.Delete(listCmd[1])
			// io.WriteString(conn, "OK\n")
		default:
			io.WriteString(conn, "Error Unknown command\n")
			continue
		}
		sA := c.Ring.GetNode(key)
		workerConn, err := net.Dial("tcp", sA)
		if err != nil {
			io.WriteString(conn, "Error: Worker unavailable\n")
			continue
		}
		// defer workerConn.Close()

		fmt.Fprintf(workerConn, "%s\n", cmd)

		resp, err := bufio.NewReader(workerConn).ReadString('\n')
		if err != nil {
			io.WriteString(conn, "Error: Worker response failed\n")
			continue
		}

		io.WriteString(conn, resp)
		workerConn.Close()
	}
}

func (c *Coordinator) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
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
