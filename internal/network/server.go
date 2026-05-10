package network

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"nexo/internal/resp"
	"nexo/internal/store"
	"strings"
	"sync"
)

type Server struct {
	St   *store.Store[string]
	Port int
}

func (s *Server) handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer conn.Close()
	defer wg.Done()
	// sc := bufio.NewScanner(conn)
	r := bufio.NewReader(conn)

	for {
		first, err := r.Peek(1)
		if err != nil {
			break
		}
		if first[0] == '*' {
			val, err := resp.Read(r)
			if err != nil {
				break
			}
			if len(val.Array) < 1 {
				continue
			}
			cmd := val.Array[0].Str
			args := val.Array[1:]
			switch cmd {
			case "SET":
				if len(args) < 2 {
					resp.Value{Kind: '-', Str: "ERR wrong number of arguments"}.Write(conn)
					continue
				}
				s.St.Set(args[0].Str, args[1].Str)
				resp.Value{Kind: '+', Str: "OK"}.Write(conn)
			case "GET":
				if len(args) < 1 {
					resp.Value{Kind: '-', Str: "ERR wrong number of arguments"}.Write(conn)
					continue
				}
				val, ok := s.St.Get(args[0].Str)
				if ok {
					resp.Value{Kind: '$', Str: val}.Write(conn)
				} else {
					resp.Value{Kind: '-', Str: "key not found"}.Write(conn)
				}
			case "DEL":
				if len(args) < 1 {
					resp.Value{Kind: '-', Str: "ERR wrong number of arguments"}.Write(conn)
					continue
				}
				s.St.Delete(args[0].Str)
				resp.Value{Kind: '+', Str: "OK"}.Write(conn)
			default:
				resp.Value{Kind: '-', Str: "ERR unknown command"}.Write(conn)
			}

		} else {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			cmd := strings.TrimRight(line, "\r\n")
			// cmd := sc.Text()

			listCmd := strings.Fields(cmd)
			if len(listCmd) == 0 {
				continue
			}
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
}

func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
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
		go s.handleConnection(conn, wg)

	}
}
