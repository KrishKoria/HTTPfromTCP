package response

import (
	"fmt"
	"io"

	"github.com/KrishKoria/HTTPfromTCP/internal/headers"
)

type StatusCode int

const (
    StatusOK StatusCode = 200
    
    StatusBadRequest StatusCode = 400
    
    StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reasonPhrase string
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		return fmt.Errorf("unknown status code: %d", statusCode)
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
    h := headers.NewHeaders()
    h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
    h.Set("Connection", "close")
    h.Set("Content-Type", "text/plain")
    return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for name, value := range headers {
        headerLine := fmt.Sprintf("%s: %s\r\n", name, value)
        _, err := w.Write([]byte(headerLine))
        if err != nil {
            return err
        }
    }
    
    _, err := w.Write([]byte("\r\n"))
    return err
}