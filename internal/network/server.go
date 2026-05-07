package network

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"nexo/internal/store"
	"strings"
)

type Server struct {
	St   *store.Store[string]
	Port int
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	sc := bufio.NewScanner(conn)

	for sc.Scan() {
		cmd := sc.Text()

		listCmd := strings.Fields(cmd)

		switch listCmd[0] {
		case "SET":
			if len(listCmd) < 3 {
				io.WriteString(conn, "Set command needs key and value to work\n")
				continue
			}
			s.St.Set(listCmd[1], listCmd[2])
			io.WriteString(conn, "OK\n")
		case "GET":
			if len(listCmd) < 2 {
				io.WriteString(conn, "Get command needs key to work\n")
				continue
			}
			val, ok := s.St.Get(listCmd[1])
			if ok {
				fmt.Fprintf(conn, "%s\n", val)
			} else {
				io.WriteString(conn, "Error key not found\n")
			}
		case "DEL":
			if len(listCmd) < 2 {
				io.WriteString(conn, "Delete command needs key to work\n")
				continue
			}
			s.St.Delete(listCmd[1])
			io.WriteString(conn, "OK\n")
		default:
			io.WriteString(conn, "Error Unknown command\n")
			continue
		}

	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return
	}
	for {
		conn, err := listener.Accept()

		if err != nil {
			return
		}

		go s.handleConnection(conn)

	}
}
