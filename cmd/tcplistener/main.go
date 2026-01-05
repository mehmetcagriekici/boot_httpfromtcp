package main

import(
	"log"
	"fmt"
	"net"
	// "io"
        // "strings"

	"github.com/mehmetcagriekici/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Println("A connection has been accepted.")
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		// lines := getLinesChannel(conn)
		// for line := range lines {
		//	fmt.Println(line)
		// }
	}	
}

/*
func getLinesChannel(conn net.Conn) <-chan string {
	buf := make([]byte, 8)
	currLine := ""
	result := make(chan string)
        go func() <- chan string {
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					if currLine != "" {
						result <- currLine
					}
					conn.Close()
					close(result)
					return result
				}
				log.Fatal(err)
			}
			parts := strings.Split(string(buf[:n]), "\n")
			for i, part := range parts {
				if i < len(parts) - 1 {
					result <- currLine + part
					currLine = ""
				} else {
					currLine += part
				}
			}
		}
		return result
	}()
	return result
}
*/
