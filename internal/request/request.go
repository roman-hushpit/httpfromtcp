package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/roman-hushpit/learn-http-protocol/internal/headers"
)

type State int

const (
	requestStateInitialized State = iota
	requestStateParsingHeaders
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       State
}

func (req Request) String() string {
	return fmt.Sprintf("Request line:\n%s\nHeaders:\n%s", req.RequestLine, req.Headers)
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (rl RequestLine) String() string {
	return fmt.Sprintf("- Method: %s\n- Target: %s\n- Version: %s", rl.Method, rl.RequestTarget, rl.HttpVersion)
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := new(Request)
	request.Headers = headers.NewHeaders()

	buf := make([]byte, bufferSize)
	readToIndex := 0
	for request.State != requestStateDone {
		byteReaded, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.State = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += byteReaded

		parsedBytes, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if parsedBytes == 0 && readToIndex == len(buf) {
			bigger := make([]byte, len(buf)*2)
			copy(bigger, buf)
			buf = bigger
			continue
		} else if parsedBytes == 0 {
			continue
		} else {
			copy(buf, buf[parsedBytes:readToIndex])
			readToIndex -= parsedBytes
			continue
		}
	}
	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.State != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case requestStateInitialized:
		bytesRead, requestLine, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if requestLine == nil {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.State = requestStateParsingHeaders
		return bytesRead, nil
	case requestStateParsingHeaders:
		bytesRead, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.State = requestStateDone
		}
		return bytesRead, nil
	default:
		return 0, fmt.Errorf("invalid request state: %d", r.State)
	}
}

func parseRequestLine(input []byte) (int, *RequestLine, error) {
	contains := bytes.Contains(input, []byte(crlf))
	if !contains {
		return 0, nil, nil
	}
	requestLineBytes := bytes.Split(input, []byte(crlf))[0]
	line := string(requestLineBytes)
	requestLineArguments := strings.Split(line, " ")
	if len(requestLineArguments) != 3 {
		return 0, nil, fmt.Errorf("invalid request line: %s", line)
	}
	method := requestLineArguments[0]

	if matched, _ := regexp.Match("^([A-Z]+)$", []byte(method)); !matched {
		return 0, nil, fmt.Errorf("invalid method name: %s", line)
	}

	target := requestLineArguments[1]

	httpVersion := requestLineArguments[2]
	versionArguments := strings.Split(httpVersion, "/")
	if len(versionArguments) != 2 || versionArguments[1] != "1.1" {
		return 0, nil, fmt.Errorf("invalid http version: %s", httpVersion)
	}
	return len(requestLineBytes) + 2, &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   versionArguments[1],
	}, nil
}
