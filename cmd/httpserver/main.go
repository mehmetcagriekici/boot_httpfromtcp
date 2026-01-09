package main

import(
	"io"
	"log"
	"fmt"
	"os"
	"syscall"
	"os/signal"
	"strings"
	"net/http"
	"errors"
	"crypto/sha256"
	
	"github.com/mehmetcagriekici/httpfromtcp/internal/server"
	"github.com/mehmetcagriekici/httpfromtcp/internal/request"
	"github.com/mehmetcagriekici/httpfromtcp/internal/response"
	"github.com/mehmetcagriekici/httpfromtcp/internal/headers"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func proxyHandler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget
	url := fmt.Sprintf("https://httpbin.org%s", strings.TrimPrefix(target, "/httpbin"))
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	statusCode := response.STATUS_INTERNAL_SERVER_ERROR
	if resp.StatusCode == 200 {
		statusCode = response.STATUS_OK
	}
	if resp.StatusCode == 400 {
		statusCode = response.STATUS_BAD_REQUEST
	}
	if err := w.WriteStatusLine(statusCode); err != nil {
		log.Fatal(err)
	}

	bheaders := headers.Headers{
		"Connection":        "close",
		"Content-Type":      "application/json",
		"Transfer-Encoding": "chunked",
		"Trailer":           "X-Content-SHA256, X-Content-Length",
	}
	if err := w.WriteHeaders(bheaders); err != nil {
		log.Fatal(err)
	}

	bufLen := 1024
	buf := make([]byte, bufLen, bufLen)
	totalLength := 0
	fullBody := []byte{}
	for {
		nr, err := resp.Body.Read(buf)
		if err != nil {
		        if errors.Is(err, io.EOF) {
				w.WriteChunkedBodyDone()
				break
			}
			log.Fatal(err)
		}
		chunk := fmt.Sprintf("%x\r\n%s\r\n", nr, string(buf[:nr]))
		nw, err := w.WriteChunkedBody([]byte(chunk))
		if err != nil {
			log.Println(err)
			break
		}
		totalLength += nr
		fullBody = append(fullBody, buf[:nr]...)
		log.Printf("Bytes written: %d\n", nw)
	}

	trailerHeaders := headers.Headers{
		"X-Content-Length": fmt.Sprintf("%d", totalLength),
		"X-Content-SHA256": fmt.Sprintf("%x", sha256.Sum256(fullBody)),
	}
	if err := w.WriteTrailers(trailerHeaders); err != nil {
		log.Fatal(err)
	}
	log.Printf("Trailer X-Content-Length=%s", trailerHeaders["X-Content-Length"])
	log.Printf("Trailer X-Content-SHA256=%s", trailerHeaders["X-Content-SHA256"])
}

func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget
	method := req.RequestLine.Method
	if strings.HasPrefix(target, "/httpbin") {
		proxyHandler(w, req)
		return
	}

	if method == "GET" && target == "/video" {
		if err := w.WriteStatusLine(response.STATUS_OK); err != nil {
			log.Fatal(err)
		}
		b, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Fatal(err)
		}
		headers := headers.Headers{
			"Content-Type": "video/mp4",
			"Connection": "close",
			"Content-Length": fmt.Sprintf("%d", len(b)),
		}
		if err := w.WriteHeaders(headers); err != nil {
			log.Fatal(err)
		}
		if _, err := w.WriteBody(b); err != nil {
			log.Fatal(err)
		}
		return
	}
	
	if target == "/yourproblem" {
		if err := w.WriteStatusLine(response.STATUS_BAD_REQUEST); err != nil {
			log.Fatal(err)
		}
		html := response.ResponseHTML("400 Bad Request", "Bad Request", "Your request honestly kinda sucked.")
		headers := response.GetDefaultHeaders(len(html))
                if err := w.WriteHeaders(headers); err != nil {
			log.Fatal(err)
		}
		if _, err := w.WriteBody([]byte(html)); err != nil {
			log.Fatal(err)
		}
 	} else if target == "/myproblem" {
		if err := w.WriteStatusLine(response.STATUS_INTERNAL_SERVER_ERROR); err != nil {
			log.Fatal(err)
		}
		html := response.ResponseHTML("500 Internal Server Error", "Internal Server Error", "Okay, you know what? This one is on me.")
		headers := response.GetDefaultHeaders(len(html))
                if err := w.WriteHeaders(headers); err != nil {
			log.Fatal(err)
		}
		if _, err := w.WriteBody([]byte(html)); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := w.WriteStatusLine(response.STATUS_OK); err != nil {
			log.Fatal(err)
		}
		html := response.ResponseHTML("200 OK", "Success!", "Your request was an absolute banger.")
		headers := response.GetDefaultHeaders(len(html))
		if err := w.WriteHeaders(headers); err != nil {
			log.Fatal(err)
		}
		if _, err := w.WriteBody([]byte(html)); err != nil {
			log.Fatal(err)
		}
	}
}
