package main

import(
	"log"
	"os"
	"syscall"
	"os/signal"
	
	"github.com/mehmetcagriekici/httpfromtcp/internal/server"
	"github.com/mehmetcagriekici/httpfromtcp/internal/request"
	"github.com/mehmetcagriekici/httpfromtcp/internal/response"
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

func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget
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
