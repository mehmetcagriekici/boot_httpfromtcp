package main

import(
	"log"
	"os"
	"io"
        "strings"
)

func main() {
	messagesFile, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal(err)
	}
	lines := getLinesChannel(messagesFile)
	for line := range lines {
		log.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	buf := make([]byte, 8)
	currLine := ""
	result := make(chan string)
        go func() <- chan string {
		for {
			n, err := f.Read(buf)
			if err != nil {
				if err == io.EOF {
					if currLine != "" {
						result <- currLine
					}
					f.Close()
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
