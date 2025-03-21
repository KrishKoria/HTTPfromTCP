package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestFromReader(t *testing.T) {
    // Test: Good GET Request line
    reader := &chunkReader{
        data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
        numBytesPerRead: 3,
    }
    r, err := RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "GET", r.RequestLine.Method)
    assert.Equal(t, "/", r.RequestLine.RequestTarget)
    assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)

    // Test: Good GET Request line with path
    reader = &chunkReader{
        data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
        numBytesPerRead: 1,
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "GET", r.RequestLine.Method)
    assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
    assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)

    // Test: Good POST Request with path
    reader = &chunkReader{
        data:            "POST /submit HTTP/1.1\r\nContent-Type: application/json\r\n\r\n{\"data\":\"test\"}",
        numBytesPerRead: 5,
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "POST", r.RequestLine.Method)
    assert.Equal(t, "/submit", r.RequestLine.RequestTarget)
    assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)

    // Test: Invalid number of parts in request line
    reader = &chunkReader{
        data:            "GET /index HTTP\r\nHost: localhost\r\n\r\n",
        numBytesPerRead: 4,
    }
    r, err = RequestFromReader(reader)
    require.Error(t, err)
    require.Nil(t, r)

    // Test: Invalid method (out of order) Request line
    reader = &chunkReader{
        data:            "get / HTTP/1.1\r\nHost: localhost\r\n\r\n",
        numBytesPerRead: 2,
    }
    r, err = RequestFromReader(reader)
    require.Error(t, err)
    require.Nil(t, r)

    // Test: Invalid version in Request line
    reader = &chunkReader{
        data:            "GET / HTTP/2.0\r\nHost: localhost\r\n\r\n",
        numBytesPerRead: 7,
    }
    r, err = RequestFromReader(reader)
    require.Error(t, err)
    require.Nil(t, r)

    // Test with different chunk sizes
    reader = &chunkReader{
        data:            "GET /api/v1/users HTTP/1.1\r\nHost: api.example.com\r\n\r\n",
        numBytesPerRead: 10, // Larger chunk size
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "GET", r.RequestLine.Method)
    assert.Equal(t, "/api/v1/users", r.RequestLine.RequestTarget)
    assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)

    // Test with a chunk size of exactly one request line
    reader = &chunkReader{
        data:            "PUT /resource HTTP/1.1\r\nContent-Type: text/plain\r\n\r\n",
        numBytesPerRead: 22, // Exactly the length of "PUT /resource HTTP/1.1"
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "PUT", r.RequestLine.Method)
    assert.Equal(t, "/resource", r.RequestLine.RequestTarget)
    assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)
    
    // Test with the entire request read at once
    fullRequest := "DELETE /items/42 HTTP/1.1\r\nAuthorization: Bearer token123\r\n\r\n"
    reader = &chunkReader{
        data:            fullRequest,
        numBytesPerRead: len(fullRequest), // Read the entire request in one chunk
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "DELETE", r.RequestLine.Method)
    assert.Equal(t, "/items/42", r.RequestLine.RequestTarget)
    assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)
}

func TestRequestHeader(t *testing.T) {
    // Test: Standard Headers
    reader := &chunkReader{
        data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
        numBytesPerRead: 3,
    }
    r, err := RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "localhost:42069", r.Headers["host"])
    assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
    assert.Equal(t, "*/*", r.Headers["accept"])
    
    // Test: Malformed Header
    reader = &chunkReader{
        data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
        numBytesPerRead: 3,
    }
    _, err = RequestFromReader(reader)
    require.Error(t, err)
    
    // Test: Empty Headers
    reader = &chunkReader{
        data:            "GET / HTTP/1.1\r\n\r\n",
        numBytesPerRead: 5,
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, 0, len(r.Headers))
    
    // Test: Duplicate Headers
    reader = &chunkReader{
        data:            "GET / HTTP/1.1\r\nAccept: text/html\r\nAccept: application/json\r\n\r\n",
        numBytesPerRead: 4,
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "text/html,application/json", r.Headers["accept"])
    
    // Test: Case Insensitive Headers
    reader = &chunkReader{
        data:            "GET / HTTP/1.1\r\nHost: example.com\r\nHOST: example2.com\r\n\r\n",
        numBytesPerRead: 6,
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    // Headers should be case-insensitive, so the latter one overwrites or gets concatenated
    assert.Contains(t, r.Headers, "host")
    assert.True(t, r.Headers["host"] == "example.com,example2.com" || r.Headers["host"] == "example2.com")
    
    // Test: Missing End of Headers (should eventually timeout/error in real-world)
    // Here we're using a reader that will simulate complete data with EOF
    reader = &chunkReader{
        data:            "GET / HTTP/1.1\r\nHost: example.com\r\n",
        numBytesPerRead: 8,
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err) // Should be OK since EOF is treated as end
    require.NotNil(t, r)
    assert.Equal(t, "example.com", r.Headers["host"])
    
    // Test: Header with whitespace
    reader = &chunkReader{
        data:            "GET / HTTP/1.1\r\nContent-Type: application/json; charset=utf-8\r\n\r\n",
        numBytesPerRead: 7,
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "application/json; charset=utf-8", r.Headers["content-type"])
}

type chunkReader struct {
    data            string
    numBytesPerRead int
    pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
    if cr.pos >= len(cr.data) {
        return 0, io.EOF
    }
    endIndex := cr.pos + cr.numBytesPerRead
    if endIndex > len(cr.data) {
        endIndex = len(cr.data)
    }
    n = copy(p, cr.data[cr.pos:endIndex])
    cr.pos += n
    if n > cr.numBytesPerRead {
        n = cr.numBytesPerRead
        cr.pos -= n - cr.numBytesPerRead
    }
    return n, nil
}