package response

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/roman-hushpit/learn-http-protocol/internal/headers"
)

type StatusCode int

const (
	SUCCESS               StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	INTERNAL_SERVER_ERROR StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case SUCCESS:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
		return nil
	case BAD_REQUEST:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
		return nil
	case INTERNAL_SERVER_ERROR:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
		return nil
	default:
		return nil
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Type"] = "text/plain"
	headers["Content-Length"] = strconv.Itoa(contentLen)
	headers["Connection"] = "close"
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

type WriterStatus int

const (
	nothingWritten WriterStatus = iota
	statusLineWritten
	headersLineWritten
	bodyWritten
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}

type Writer struct {
	writer       io.Writer
	WriterStatus WriterStatus
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.WriterStatus == nothingWritten {
		err := WriteStatusLine(w.writer, statusCode)
		if err != nil {
			return err
		}
		w.WriterStatus = statusLineWritten
		return nil
	}
	return errors.New("wrong order of writing")
}
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.WriterStatus == statusLineWritten {
		err := WriteHeaders(w.writer, headers)
		if err != nil {
			return err
		}
		w.WriterStatus = headersLineWritten
		return nil
	}
	return errors.New("wrong order of writing")

}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.WriterStatus == headersLineWritten {
		w.writer.Write([]byte(fmt.Sprintf("%x\r\n", 0)))
		WriteHeaders(w.writer, h)
		w.writer.Write([]byte("\r\n"))
		w.WriterStatus = bodyWritten
	}
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.WriterStatus == headersLineWritten {
		n, err := w.writer.Write(p)
		w.WriterStatus = bodyWritten
		return n, err
	}
	return 0, errors.New("wrong order of writing")
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.WriterStatus == headersLineWritten {
		total := 0
		n, _ := w.writer.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
		total += n
		n, _ = w.writer.Write(p)
		total += n
		n, _ = w.writer.Write([]byte("\r\n"))
		total += n
		return total, nil
	}
	return 0, errors.New("wrong order of writing")
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.WriterStatus == headersLineWritten {
		total := 0
		n, _ := w.writer.Write([]byte(fmt.Sprintf("%x\r\n", 0)))
		total += n
		n, _ = w.writer.Write([]byte("\r\n"))
		total += n
		w.WriterStatus = bodyWritten
		return total, nil
	}
	return 0, errors.New("wrong order of writing")
}
