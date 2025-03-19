package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
        
        lines := getLinesChannel(conn)
        for line := range lines {
            fmt.Println(line)
        }
        
        log.Printf("Connection closed")
    }
}

func getLinesChannel(f io.ReadCloser) <-chan string {
    lines := make(chan string)
    go func() {
        defer f.Close()
        
        buffer := make([]byte, 8)
        currentLine := ""
        
        for {
            bytesRead, err := f.Read(buffer)
            if err == io.EOF {
                break
            }
            if err != nil {
                log.Fatal(err)
            }
            
            data := string(buffer[:bytesRead])
            parts := strings.Split(data, "\n")
            
            for i := 0; i < len(parts); i++ {
                if i < len(parts)-1 {
                    currentLine += parts[i]
                    lines <- currentLine
                    currentLine = ""
                } else if strings.HasSuffix(data, "\n") {
                    currentLine += parts[i]
                    lines <- currentLine
                    currentLine = ""
                } else {
                    currentLine += parts[i]
                }
            }
        }
        
        if currentLine != "" {
            lines <- currentLine
        }
        
        close(lines)
    }()
    
    return lines
}