package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/roman-hushpit/learn-http-protocol/internal/headers"
)

type State int

const (
	requestStateInitialized State = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       State
}

func (req Request) String() string {
	return fmt.Sprintf("Request line:\n%s\nHeaders:\n%sBody:\n%s\n", req.RequestLine, req.Headers, req.Body)
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (rl RequestLine) String() string {
	return fmt.Sprintf("- Method: %s\n- Target: %s\n- Version: %s",
		rl.Method, rl.RequestTarget, rl.HttpVersion)
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := new(Request)
	request.Headers = headers.NewHeaders()

	buf := make([]byte, bufferSize)
	readToIndex := 0
	for request.State != requestStateDone {
		byteRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.State != requestStateDone {
					// still parsing body but stream ended → too short
					return nil, errors.New("incomplete body")
				}
				return request, nil
			}
			return nil, err
		}
		readToIndex += byteRead

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
			r.State = requestStateParsingBody
		}
		return bytesRead, nil
	case requestStateParsingBody:
		if r.Headers.Get("Content-Length") == "" {
			r.State = requestStateDone
			return 0, nil
		}
		r.Body = append(r.Body, data...)
		err := r.checkBodyProgress()
		if r.State == requestStateDone {
			return 0, err
		}

		if err != nil {
			return len(data), err
		}
		return len(data), nil
	default:
		return 0, fmt.Errorf("invalid request state: %d", r.State)
	}
}

func (r *Request) checkBodyProgress() error {
	contentLength, ok, err := r.contentLength()
	if err != nil {
		return err
	}
	if !ok {
		r.State = requestStateDone
		return nil
	}
	if len(r.Body) > contentLength {
		return errors.New("body too large")
	}
	if len(r.Body) == contentLength {
		r.State = requestStateDone
	}
	return nil
}

func (r Request) contentLength() (int, bool, error) {
	contentLengthHeader := r.Headers.Get("Content-Length")
	if contentLengthHeader == "" || contentLengthHeader == "0" {
		return 0, false, nil
	}
	contentLength, err := strconv.Atoi(contentLengthHeader)
	if err != nil {
		return 0, false, err
	}
	return contentLength, true, nil
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
