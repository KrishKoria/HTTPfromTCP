package main

import (
	"io"
	"log"
	"os"
	"os/signal"
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
