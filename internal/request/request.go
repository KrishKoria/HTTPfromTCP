package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	state RequestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type RequestState int

const (
	requestStateInitialized RequestState = iota
	requestStateDone
)
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error){
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		state: requestStateInitialized,
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
		r.state = requestStateDone
		return bytesUsed, nil
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