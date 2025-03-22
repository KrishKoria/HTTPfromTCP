package server

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/KrishKoria/HTTPfromTCP/internal/response"
)

type Server struct {
	listener net.Listener
	closed atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error listening: %w", err)
	}
	server := &Server{listener: listener}
	go server.listen()

	return server, nil
}

func(s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
	return s.listener.Close()
	}
	return nil
}

func(s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go s.handle(conn)
	}
}

func(s *Server) handle(conn net.Conn) {
	defer conn.Close()
	err := response.WriteStatusLine(conn, response.StatusOK)
    if err != nil {
        fmt.Printf("Error writing status line: %v\n", err)
        return
    }
    
    headers := response.GetDefaultHeaders(0)
    
    err = response.WriteHeaders(conn, headers)
    if err != nil {
        fmt.Printf("Error writing headers: %v\n", err)
        return
    }
}
