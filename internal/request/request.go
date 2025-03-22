package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/KrishKoria/HTTPfromTCP/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	state RequestState
	Headers headers.Headers
	Body []byte
	BodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type RequestState int

const (
	requestStateInitialized RequestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error){
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		state: requestStateInitialized,
		Headers: headers.NewHeaders(),
		Body: make([]byte, 0),
	}
	for req.state != requestStateDone {
		if(readToIndex >= len(buf)){
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}
		bytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += bytesRead

		bytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[bytesParsed:])
		readToIndex -= bytesParsed
	}
	return req, nil
}

func(r *Request) parse(data []byte) (int, error){
	totalBytesParsed := 0
    
    for r.state != requestStateDone {
        bytesUsed, err := r.parseSingle(data[totalBytesParsed:])
        if err != nil {
            return 0, err
        }
        if bytesUsed == 0 {
            break
        }
        totalBytesParsed += bytesUsed
        if totalBytesParsed >= len(data) {
            break
        }
    }
    
    return totalBytesParsed, nil
}
func(r *Request) parseSingle(data []byte) (int, error){
	switch r.state {
	case requestStateInitialized:
		requestLine, bytesUsed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if bytesUsed == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return bytesUsed, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLengthStr, ok := r.Headers.Get("content-length")
		if !ok {
			r.state = requestStateDone
			return len(data), nil
		}
		
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid Content-Length: %w", err)
		}
		r.Body = append(r.Body, data...)
		r.BodyLengthRead += len(data)
		if r.BodyLengthRead > contentLength {
			return 0, fmt.Errorf("Content-Length too large")
		}
		if r.BodyLengthRead == contentLength {
			r.state = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, fmt.Errorf("request is already done")
	default:
		return 0, fmt.Errorf("unknown request state")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error){
 	lines := strings.Split(str, "\r\n")
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
 	return &RequestLine{HttpVersion: httpVersion, RequestTarget: requestTarget, Method: method}, nil
}