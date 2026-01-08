package server

import(
	"fmt"
	"log"
	"net"
	"sync/atomic"
	
	"github.com/mehmetcagriekici/httpfromtcp/internal/response"
	"github.com/mehmetcagriekici/httpfromtcp/internal/request"
)

type HandlerError struct{
	StatusCode response.StatusCode
	Message string
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct{
	Listener net.Listener
	handler Handler
	b atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		Listener: l,
		handler: handler,
	}
	go s.Listen()
	return s, nil
}

func (s *Server) Close() error {
	s.b.Store(true)
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

func (s *Server) Listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.b.Load() {
				return
			}
			log.Println(err)
			continue
		}
		go s.Handle(conn)
	}
}

func (s *Server) Handle(conn net.Conn) {
	defer conn.Close()
	w := &response.Writer{
		Writer: conn,
		WriteState: response.STATE_WRITE_STATUS_LINE,
	}

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println(err)
		return
	}
	s.handler(w, req)
}

func WriteError(w response.Writer, herr *HandlerError) {
	if err := w.WriteStatusLine(herr.StatusCode); err != nil {
		log.Println(err)
		return
	}
	headers := response.GetDefaultHeaders(len(herr.Message))
	if err := w.WriteHeaders(headers); err != nil {
		log.Println(err)
		return
	}
	if _, err := w.WriteBody([]byte(herr.Message)); err != nil {
		log.Println(err)
		return
	}
	return
}
