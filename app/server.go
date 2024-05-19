package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buffer := make([]byte, 1024)
	conn.Read(buffer)
	request := string(buffer)

	targetRegex := regexp.MustCompile(`^GET (.*) HTTP/1\.1`)
	target := targetRegex.FindStringSubmatch(request)[1]

	if target == "/" {
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
	}

	if strings.HasPrefix(target, "/echo/") {
		echoRegex := regexp.MustCompile(`/echo/(.*)`)
		echo := echoRegex.FindStringSubmatch(target)

		if echo != nil {
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-type: text/plain\r\nContent-Length: %d\r\n\r\n%s", utf8.RuneCountInString(echo[1]), echo[1])
		}
		fmt.Fprintf(conn, "HTTP/1.1 400 BAD REQUEST\r\n\r\n")
	}

	if target == "/user-agent" {
		userAgentRegex := regexp.MustCompile(`User-Agent: (\S+)`)
		userAgent := userAgentRegex.FindStringSubmatch(request)[1]
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", utf8.RuneCountInString(userAgent), userAgent)
	}

	fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
}
