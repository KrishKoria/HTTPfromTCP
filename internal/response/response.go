package response

import (
	"fmt"
	"io"

	"github.com/KrishKoria/HTTPfromTCP/internal/headers"
)

type StatusCode int
type writerState int

const (
    writerStateInitialized writerState = iota
    writerStateStatusLineWritten
    writerStateHeadersWritten
    writerStateBodyWritten
)

type Writer struct {
    w     io.Writer
    state writerState
}
const (
    StatusOK StatusCode = 200
    
    StatusBadRequest StatusCode = 400
    
    StatusInternalServerError StatusCode = 500
)

func NewWriter(w io.Writer) *Writer {
    return &Writer{
        w:     w,
        state: writerStateInitialized,
    }
}

func getStatusLine(statusCode StatusCode) []byte {
	reasonPhrase := ""
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase))
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
    if w.state != writerStateInitialized {
        return fmt.Errorf("status line must be written first (current state: %d)", w.state)
    }
    _, err := w.w.Write(getStatusLine(statusCode))
    if err != nil {
        return err
    }
    w.state = writerStateStatusLineWritten
    return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
    if w.state != writerStateStatusLineWritten {
        return fmt.Errorf("headers must be written after status line (current state: %d)", w.state)
    }
    for k, v := range headers {
        _, err := w.w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
        if err != nil {
            return err
        }
    }
    _, err := w.w.Write([]byte("\r\n"))
    if err != nil {
        return err
    }
    w.state = writerStateHeadersWritten
    return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
    if w.state != writerStateHeadersWritten {
        return 0, fmt.Errorf("body must be written after headers (current state: %d)", w.state)
    }
    n, err := w.w.Write(p)
    if err != nil {
        return n, err
    }
    w.state = writerStateBodyWritten
    return n, nil
}