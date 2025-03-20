package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error){
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}
	requestLine := string(data)
	lines := strings.Split(requestLine, "\r\n")
	if len(lines) == 0 || lines[0] == "" {
        return nil, fmt.Errorf("empty request")
    }
	parts := strings.Split(lines[0], " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid number of parts in request line")
	}
	method := parts[0]
	requestTarget := parts[1]
	httpVersion := parts[2]
	for _, ch := range method {
		if ch < 'A' || ch > 'Z' {
			return nil, fmt.Errorf("invalid method : %s", method)
		}
	}
	if httpVersion != "HTTP/1.1" {
		return nil, fmt.Errorf("invalid HTTP version : %s", httpVersion)
	}
	return &Request{
		RequestLine: RequestLine{
			HttpVersion:   httpVersion,
			RequestTarget: requestTarget,
			Method:        method,
		},
	}, nil
}