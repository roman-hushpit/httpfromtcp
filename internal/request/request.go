package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Status int

const (
	INITIALIZED Status = iota
	DONE
)

type Request struct {
	RequestLine RequestLine
	Status      Status
}

func (req Request) String() string {
	return fmt.Sprintf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := new(Request)
	buf := make([]byte, bufferSize)
	readToIndex := 0
	for request.Status != DONE {
		byteReaded, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.Status = DONE
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
	bytesRead, requestLine, err := parseRequestLine(data)
	if err != nil {
		return 0, err
	}
	if requestLine == nil {
		return 0, nil
	}
	r.RequestLine = *requestLine
	r.Status = DONE
	return bytesRead, nil
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
