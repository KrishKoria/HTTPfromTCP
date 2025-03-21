package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := strings.ToLower(string(parts[0]))

	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}
	
	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)
	for _, r := range key {
        if !((r >= 'A' && r <= 'Z') || 
            (r >= 'a' && r <= 'z') || 
            (r >= '0' && r <= '9') || 
            r == '!' || r == '#' || r == '$' || r == '%' || r == '&' || 
            r == '\'' || r == '*' || r == '+' || r == '-' || r == '.' || 
            r == '^' || r == '_' || r == '`' || r == '|' || r == '~') {
            return 0, false, fmt.Errorf("invalid character in header name: %s", key)
        }
    }

	if existing, ok := h[key]; ok {
		value = []byte(existing + "," + string(value))
		h.Set(key, string(value))
	} else {
		h.Set(key, string(value))
	}

	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	h[key] = value
}
