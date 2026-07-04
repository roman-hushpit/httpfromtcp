package headers

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const crlf = "\r\n"

var headerNameRE = regexp.MustCompile(`^[!#$%&'*+\-.^_` + "`" + `|~0-9A-Za-z]+$`)

type Headers map[string]string

func (h Headers) String() string {
	headers := ""
	for hk, hv := range h {
		headers += fmt.Sprintf("- %s: %s\n", hk, hv)
	}
	return headers
}

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Set(key string, value string) {
	h[key] = strings.TrimSpace(value)
}

func (h Headers) Drop(key string) {
	delete(h, key)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfPosition := bytes.Index(data, []byte(crlf))
	if crlfPosition == -1 {
		return 0, false, nil
	}
	if crlfPosition == 0 {
		return 2, true, nil
	}
	headerLine := string(data[:crlfPosition])
	headerValues := strings.SplitN(headerLine, ":", 2)
	if len(headerValues) != 2 {
		return 0, false, errors.New("invalid header line")
	}
	headerName := headerValues[0]
	headerValue := headerValues[1]
	if strings.ContainsAny(headerName, " \t") {
		return 0, false, errors.New("invalid header line")
	}
	if !headerNameRE.MatchString(headerName) {
		return 0, false, errors.New("invalid header line")
	}
	headerName = strings.ToLower(headerName)
	headerValue = strings.TrimSpace(headerValue)
	if _, ok := h[headerName]; ok {
		h[headerName] = h[headerName] + ", " + headerValue
	} else {
		h[headerName] = headerValue
	}
	return crlfPosition + 2, false, nil
}
