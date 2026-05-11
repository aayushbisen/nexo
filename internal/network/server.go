package network

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"nexo/internal/resp"
	"nexo/internal/store"
	"strconv"
	"strings"
	"sync"
	"time"
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
				key := args[0].Str
				value := args[1].Str
				var ttl time.Duration
				for i := 2; i < len(args); i++ {
					switch strings.ToUpper(args[i].Str) {
					case "EX":
						if i+1 < len(args) {
							sec, err := strconv.Atoi(args[i+1].Str)
							if err == nil {
								ttl = time.Duration(sec) * time.Second
							}
							i++
						}
					case "PX":
						if i+1 < len(args) {
							ms, err := strconv.Atoi(args[i+1].Str)
							if err == nil {
								ttl = time.Duration(ms) * time.Millisecond
							}
							i++
						}
					}
				}
				if ttl > 0 {
					s.St.Set(key, value, ttl)
				} else {
					s.St.Set(key, value)
				}
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
			case "EXPIRE":
				if len(args) < 2 {
					resp.Value{Kind: '-', Str: "ERR wrong number of arguments"}.Write(conn)
					continue
				}
				sec, err := strconv.Atoi(args[1].Str)
				if err != nil {
					resp.Value{Kind: '-', Str: "ERR value is not an integer or out of range"}.Write(conn)
					continue
				}
				ok := s.St.Expire(args[0].Str, time.Duration(sec)*time.Second)
				if ok {
					resp.Value{Kind: ':', Integer: 1}.Write(conn)
				} else {
					resp.Value{Kind: ':', Integer: 0}.Write(conn)
				}
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
				key := listCmd[1]
				value := listCmd[2]
				var ttl time.Duration
				for i := 3; i < len(listCmd); i++ {
					switch strings.ToUpper(listCmd[i]) {
					case "EX":
						if i+1 < len(listCmd) {
							sec, err := strconv.Atoi(listCmd[i+1])
							if err == nil {
								ttl = time.Duration(sec) * time.Second
							}
							i++
						}
					case "PX":
						if i+1 < len(listCmd) {
							ms, err := strconv.Atoi(listCmd[i+1])
							if err == nil {
								ttl = time.Duration(ms) * time.Millisecond
							}
							i++
						}
					}
				}
				if ttl > 0 {
					s.St.Set(key, value, ttl)
				} else {
					s.St.Set(key, value)
				}
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
			case "EXPIRE":
				if len(listCmd) < 3 {
					io.WriteString(conn, "Expire command needs key and seconds to work\n")
					continue
				}
				sec, err := strconv.Atoi(listCmd[2])
				if err != nil {
					io.WriteString(conn, "Error value is not an integer\n")
					continue
				}
				ok := s.St.Expire(listCmd[1], time.Duration(sec)*time.Second)
				if ok {
					io.WriteString(conn, "1\n")
				} else {
					io.WriteString(conn, "0\n")
				}
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
