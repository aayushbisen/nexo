package network

import (
	"fmt"
	"io"
	"net"
	"nexo/internal/store"
)

type Server struct {
	St   *store.Store[string]
	Port int
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	io.WriteString(conn, "Welcome to Nexo!")

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
