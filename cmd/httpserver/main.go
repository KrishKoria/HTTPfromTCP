package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/KrishKoria/HTTPfromTCP/internal/headers"
	"github.com/KrishKoria/HTTPfromTCP/internal/request"
	"github.com/KrishKoria/HTTPfromTCP/internal/response"
	"github.com/KrishKoria/HTTPfromTCP/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w io.Writer, req *request.Request) *server.HandlerError {

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
        return handleHttpbinProxy(w, req)
    }
	
    if req.RequestLine.RequestTarget == "/yourproblem" {
        return &server.HandlerError{
            StatusCode: response.StatusBadRequest,
            Message: `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`,
            Headers: headers.Headers{"Content-Type": "text/html"},
        }
    }
    
    if req.RequestLine.RequestTarget == "/myproblem" {
        return &server.HandlerError{
            StatusCode: response.StatusInternalServerError,
            Message: `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`,
            Headers: headers.Headers{"Content-Type": "text/html"},
        }
    }
    
    _, err := w.Write([]byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`))
    
    if err != nil {
        return &server.HandlerError{
            StatusCode: response.StatusInternalServerError,
            Message:    "Failed to write response",
        }
    }
    
    return nil
}

func handleHttpbinProxy(w io.Writer, req *request.Request) *server.HandlerError {
    targetPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
    httpbinURL := "https://httpbin.org" + targetPath
    
    log.Printf("Proxying request to: %s", httpbinURL)
    
    httpResp, err := http.Get(httpbinURL)
    if err != nil {
        return &server.HandlerError{
            StatusCode: response.StatusInternalServerError,
            Message:    fmt.Sprintf("Error proxying to httpbin: %v", err),
        }
    }
    defer httpResp.Body.Close()
    
    writer := response.NewWriter(w)
    
    statusCode := response.StatusOK
    if httpResp.StatusCode == 400 {
        statusCode = response.StatusBadRequest
    } else if httpResp.StatusCode == 500 {
        statusCode = response.StatusInternalServerError
    }
    
    err = writer.WriteStatusLine(statusCode)
    if err != nil {
        return &server.HandlerError{
            StatusCode: response.StatusInternalServerError,
            Message:    fmt.Sprintf("Error writing status line: %v", err),
        }
    }
    
    h := headers.NewHeaders()
    h.Set("Transfer-Encoding", "chunked")
    
    contentType := httpResp.Header.Get("Content-Type")
    if contentType != "" {
        h.Set("Content-Type", contentType)
    }
    
    h.Set("Connection", "close")
    
    err = writer.WriteHeaders(h)
    if err != nil {
        return &server.HandlerError{
            StatusCode: response.StatusInternalServerError,
            Message:    fmt.Sprintf("Error writing headers: %v", err),
        }
    }
    
    buffer := make([]byte, 1024)
    for {
        n, err := httpResp.Body.Read(buffer)
        if n > 0 {
            log.Printf("Read %d bytes from httpbin", n)
            
            _, writeErr := writer.WriteChunkedBody(buffer[:n])
            if writeErr != nil {
                log.Printf("Error writing chunk: %v", writeErr)
                return &server.HandlerError{
                    StatusCode: response.StatusInternalServerError,
                    Message:    fmt.Sprintf("Error writing response chunk: %v", writeErr),
                }
            }
        }
        
        if err == io.EOF {
            break
        }
        
        if err != nil {
            log.Printf("Error reading from httpbin: %v", err)
        }
    }
    
    err = writer.WriteChunkedBodyDone()
    if err != nil {
        log.Printf("Error finalizing chunked response: %v", err)
    }
    
    log.Printf("Successfully proxied request to httpbin")
    return nil
}
