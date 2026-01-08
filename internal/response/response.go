package response

import(
	"io"
	"fmt"

	"github.com/mehmetcagriekici/httpfromtcp/internal/headers"
)

type WriterState int
const(
	STATE_WRITE_STATUS_LINE WriterState = iota
	STATE_WRITE_HEADERS
	STATE_WRITE_BODY
)
type Writer struct{
	Writer io.Writer
	WriteState WriterState
}

type StatusCode int
const(
	STATUS_OK StatusCode = 200
	STATUS_BAD_REQUEST StatusCode = 400
	STATUS_INTERNAL_SERVER_ERROR StatusCode = 500
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.WriteState != STATE_WRITE_STATUS_LINE {
		return fmt.Errorf("Wrong writer state!")
	}
	reasonPhrase := ""
	switch statusCode {
	case STATUS_OK:
		reasonPhrase = "HTTP/1.1 200 OK\r\n"
	case STATUS_BAD_REQUEST:
		reasonPhrase = "HTTP/1.1 400 Bad Request\r\n"
	case STATUS_INTERNAL_SERVER_ERROR:
		reasonPhrase = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
	        fmt.Printf("Invalid Status Code! reason-phrase: %s\n", reasonPhrase)
	}

	if _, err := w.Writer.Write([]byte(reasonPhrase)); err != nil {
		return err
	}

	w.WriteState = STATE_WRITE_HEADERS
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": fmt.Sprintf("%d", contentLen),
		"Connection": "close",
		"Content-Type": "text/html",
	}
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.WriteState != STATE_WRITE_HEADERS {
		return fmt.Errorf("Wrong writer state")
	}
        resp := ""
	for k, v := range headers {
		resp += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	
	resp += "\r\n"
	if _, err := w.Writer.Write([]byte(resp)); err != nil {
		return err
	}
	w.WriteState = STATE_WRITE_BODY
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.WriteState != STATE_WRITE_BODY {
		return 0, fmt.Errorf("Wrong writer state")
	}
	n, err := w.Writer.Write(p)
	if err != nil {
		return 0, err
	}
	w.WriteState = STATE_WRITE_STATUS_LINE
	return n, nil
}

func ResponseHTML(title, header, message string) string{
	return fmt.Sprintf(`<html>
  <head>
    <title>%s</title>
  </head>
  <body>
    <h1>%s</h1>
    <p>%s</p>
  </body>
</html>`, title, header, message)
}
