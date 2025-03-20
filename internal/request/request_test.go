package request

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestFromReader(t *testing.T) {
    t.Run("Good GET Request line", func(t *testing.T) {
        r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
        require.NoError(t, err)
        require.NotNil(t, r)
        assert.Equal(t, "GET", r.RequestLine.Method)
        assert.Equal(t, "/", r.RequestLine.RequestTarget)
        assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)
    })

    t.Run("Good GET Request line with path", func(t *testing.T) {
        r, err := RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
        require.NoError(t, err)
        require.NotNil(t, r)
        assert.Equal(t, "GET", r.RequestLine.Method)
        assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
        assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)
    })

    t.Run("Good POST Request with path", func(t *testing.T) {
        r, err := RequestFromReader(strings.NewReader("POST /submit HTTP/1.1\r\nHost: localhost:42069\r\nContent-Type: application/json\r\nContent-Length: 18\r\n\r\n{\"data\":\"example\"}"))
        require.NoError(t, err)
        require.NotNil(t, r)
        assert.Equal(t, "POST", r.RequestLine.Method)
        assert.Equal(t, "/submit", r.RequestLine.RequestTarget)
        assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)
    })

    t.Run("Invalid number of parts in request line", func(t *testing.T) {
        _, err := RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
        require.Error(t, err)
    })

    t.Run("Invalid method (lowercase) in Request line", func(t *testing.T) {
        _, err := RequestFromReader(strings.NewReader("get / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
        require.Error(t, err)
    })

    t.Run("Invalid method (symbols) in Request line", func(t *testing.T) {
        _, err := RequestFromReader(strings.NewReader("GET@POST / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
        require.Error(t, err)
    })

    t.Run("Invalid version in Request line", func(t *testing.T) {
        _, err := RequestFromReader(strings.NewReader("GET / HTTP/2.0\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
        require.Error(t, err)
    })

    t.Run("Empty request", func(t *testing.T) {
        _, err := RequestFromReader(strings.NewReader(""))
        require.Error(t, err)
    })

    t.Run("Whitespace only request", func(t *testing.T) {
        _, err := RequestFromReader(strings.NewReader("  \r\n  \r\n"))
        require.Error(t, err)
    })

    t.Run("Request without CRLF", func(t *testing.T) {
        _, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1"))
        require.NoError(t, err) // This is a valid request
    })
}