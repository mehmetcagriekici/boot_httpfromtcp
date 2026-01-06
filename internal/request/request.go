package request

import(
	"log"
	"io"
	"strings"
	"regexp"
        "errors"
	"strconv"

	"github.com/mehmetcagriekici/httpfromtcp/internal/headers"
)

const BUFFER_SIZE int = 8

type parserState int
const(
	INITIALIZED parserState = iota
	PARSING_HEADERS
	PARSING_BODY
	DONE      
)

type Request struct {
	RequestLine RequestLine
	State parserState
	Headers headers.Headers
	Body []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

 func (r *Request) parse(data []byte) (int, error) {
	if r.State == INITIALIZED {
		n, rl, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *rl
		r.State = PARSING_HEADERS

		return n, nil
	}

	if r.State == PARSING_HEADERS {
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.State = PARSING_BODY
			if string(data) == "\r\n" {
				r.State = DONE
			}
		}
		
		return n, nil
	}

	if r.State == PARSING_BODY {
		cl, ok := r.Headers.Get("content-length")
		if !ok {
			r.State = DONE
			return 0, nil
		}

		l, err := strconv.Atoi(cl)
		if err != nil {
			log.Fatal(err)
		}

		r.Body = append(r.Body, data...)
		if len(r.Body) > l {
			return 0, errors.New("Length of body is greater than the Content-Length")
		}

		if len(r.Body) == l {
			r.State = DONE
			log.Println("Entire Body is consumed")
		}

		return len(data), nil
	}

	if r.State == DONE {		
		return 0, errors.New("Request State is Done")
	}

	return 0, errors.New("Unknown Request State")
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, BUFFER_SIZE, BUFFER_SIZE)
	readToIndex := 0

	req := &Request{
		State: INITIALIZED,
		Headers: make(headers.Headers),
		Body:    make([]byte, 0),
	}

	for {
		if req.State == DONE {
			break
		}

		if readToIndex >= len(buf) {
			newLen := len(buf) * 2
			new_buf := make([]byte, newLen, newLen)
			copy(new_buf, buf)
			buf = new_buf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.State != DONE {
					return nil, errors.New("Invalid request")
				}
				break
			}
			return nil, err
		}

		if n == 0 {
			req.State = DONE
			break
		}

		readToIndex += n
		np, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		
		copy(buf, buf[np:readToIndex])
		readToIndex -= np
	}

	if req.State != DONE {
		return nil, errors.New("Unprocessed Request")
	}

	return req, nil
}


func parseRequestLine(b []byte) (int, *RequestLine, error) {
	str := string(b)

	i := strings.Index(str, "\r\n")
	if i == -1 {
		return 0, nil, nil
	}
	
	mpv := strings.Fields(str[:i]) 
	if len(mpv) != 3 {
		return 0, nil, errors.New("Request must contain a method, path, and http version")
	}
	
	method := mpv[0]
        target := mpv[1]
	version := mpv[2]

	if matched, err := regexp.MatchString(`[^A-Z]`, method); matched || err != nil {
		return 0, nil, errors.New("Method can only contain capital alphabetic characters")
	}

	if !strings.HasPrefix(target, "/") {
		return 0, nil, errors.New("Path must start with /")
	}

	if version != "HTTP/1.1" {
		return 0, nil, errors.New("Version must be HTTP/1.1")
	}
	version = strings.Split(version, "/")[1]
	
        requestLine := &RequestLine{
		HttpVersion: version,
		Method: method,
		RequestTarget: target,
	}

	return len(str[:i+2]), requestLine, nil
}
