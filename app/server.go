package main

import (
	"flag"
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

	directory := flag.String("directory", "", "File Read Directory")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn, *directory)
	}
}

func handleRequest(conn net.Conn, directory string) {
	buffer := make([]byte, 1024)
	conn.Read(buffer)
	request := strings.TrimRight(string(buffer), "\x00")

	fmt.Println(request)

	requestRegex := regexp.MustCompile(`^(GET|POST) (.*) HTTP/1\.1`)
	parsedRequest := requestRegex.FindStringSubmatch(request)

	verb := parsedRequest[1]
	target := parsedRequest[2]

	if verb == "GET" && target == "/" {
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
	} else if verb == "GET" && strings.HasPrefix(target, "/echo/") {
		echoRegex := regexp.MustCompile(`/echo/(.*)`)
		echo := echoRegex.FindStringSubmatch(target)

		if echo != nil {
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-type: text/plain\r\nContent-Length: %d\r\n\r\n%s", utf8.RuneCountInString(echo[1]), echo[1])
		}
		fmt.Fprintf(conn, "HTTP/1.1 400 BAD REQUEST\r\n\r\n")
	} else if verb == "GET" && target == "/user-agent" {
		userAgentRegex := regexp.MustCompile(`User-Agent: (\S+)`)
		userAgent := userAgentRegex.FindStringSubmatch(request)[1]
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", utf8.RuneCountInString(userAgent), userAgent)
	} else if verb == "GET" && strings.HasPrefix(target, "/files/") {
		fileNameRegex := regexp.MustCompile(`/files/(.*)`)
		fileName := fileNameRegex.FindStringSubmatch(target)

		if fileName != nil {
			file, _ := os.ReadFile(fmt.Sprintf("%s/%s", directory, fileName[1]))
			str := string(file)

			if file != nil {
				fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", utf8.RuneCountInString(str), str)
			}
			fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
		}
	} else if verb == "POST" && strings.HasPrefix(target, "/files/") {
		fileNameRegex := regexp.MustCompile(`/files/(.*)`)
		fileName := fileNameRegex.FindStringSubmatch(target)
		bodyRegex := regexp.MustCompile(`\r\n(.*)$`)
		body := bodyRegex.FindStringSubmatch(request)

		if fileName != nil && body != nil {
			error := os.WriteFile(fmt.Sprintf("%s/%s", directory, fileName[1]), []byte(body[1]), 0666)
			fmt.Println(error)

			if error == nil {
				fmt.Fprintf(conn, "HTTP/1.1 201 Created\r\nContent-Length:%d\r\n\r\n%s", utf8.RuneCountInString(body[1]), body[1])
			}
		}
	} else {
		fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}
}
