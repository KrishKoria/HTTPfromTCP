package main

import (
	"fmt"
	"log"
	"net"

	"github.com/KrishKoria/HTTPfromTCP/internal/request"
)

func main() {
    listener, err := net.Listen("tcp", ":42069")
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()
    
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Error accepting connection: %v", err)
            continue
        }
        
        log.Printf("Connection accepted")
        
        res, err := request.RequestFromReader(conn)
        if err != nil {
            log.Printf("Error reading request: %v", err)
            continue
        }
        
		fmt.Println("Request line:")
        fmt.Printf("- Method: %s\n", res.RequestLine.Method)
        fmt.Printf("- Target: %s\n", res.RequestLine.RequestTarget)
        fmt.Printf("- Version: %s\n", res.RequestLine.HttpVersion[5:])
        fmt.Println("Headers:")
        for key, value := range res.Headers {
            fmt.Printf("- %s: %s\n", key, value)
        }

        conn.Close()
        log.Printf("Connection closed")
    }
}
