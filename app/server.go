package main

import (
	"bytes"
	"compress/gzip"
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

	registerGetHandler("", func(r request, c net.Conn) {
		fmt.Fprintf(c, "HTTP/1.1 200 OK\r\n\r\n")
	})

	registerGetHandler("user-agent", func(r request, c net.Conn) {
		userAgent, error := r.getHeader("User-Agent")

		if error != nil {
			fmt.Fprintf(c, "HTTP/1.1 400 Bad Request\r\n\r\n%s", error)
		} else {
			fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", utf8.RuneCountInString(userAgent), userAgent)
		}
	})

	registerGetHandler("echo", func(r request, c net.Conn) {
		echoRegex := regexp.MustCompile(`/echo/(.*)`)
		target, _ := r.getTarget()
		echo := echoRegex.FindStringSubmatch(target)

		acceptEncodingHeader := r.getContentEncoding()

		encodingResponseHeader := ""

		if acceptEncodingHeader != "" {
			encodingResponseHeader = fmt.Sprintf("Content-Encoding: %s\r\n", acceptEncodingHeader)

		}

		if echo != nil {

			response := echo[1]

			if acceptEncodingHeader == "gzip" {

				var buff bytes.Buffer
				zw := gzip.NewWriter(&buff)
				zw.Write([]byte(response))
				zw.Close()
				response = buff.String()
			}

			fmt.Fprintf(c, "HTTP/1.1 200 OK\r\n%sContent-type: text/plain\r\nContent-Length: %d\r\n\r\n%s", encodingResponseHeader, utf8.RuneCountInString(response), response)
		}
		fmt.Fprintf(c, "HTTP/1.1 400 BAD REQUEST\r\n\r\n")
	})

	registerGetHandler("files", func(r request, c net.Conn) {
		fileNameRegex := regexp.MustCompile(`/files/(.*)`)
		target, _ := r.getTarget()
		fileName := fileNameRegex.FindStringSubmatch(target)

		if fileName != nil {
			file, _ := os.ReadFile(fmt.Sprintf("%s/%s", *directory, fileName[1]))
			str := string(file)

			if file != nil {
				fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", utf8.RuneCountInString(str), str)
			}
			fmt.Fprintf(c, "HTTP/1.1 404 Not Found\r\n\r\n")
		}
	})

	registerPostHandler("files", func(r request, c net.Conn) {
		fileNameRegex := regexp.MustCompile(`/files/(.*)`)
		target, _ := r.getTarget()
		fileName := fileNameRegex.FindStringSubmatch(target)
		bodyRegex := regexp.MustCompile(`\r\n(.*)$`)
		body := bodyRegex.FindStringSubmatch(string(r))

		if fileName != nil && body != nil {
			error := os.WriteFile(fmt.Sprintf("%s/%s", *directory, fileName[1]), []byte(body[1]), 0666)
			fmt.Println(error)

			if error == nil {
				fmt.Fprintf(c, "HTTP/1.1 201 Created\r\nContent-Length:%d\r\n\r\n%s", utf8.RuneCountInString(body[1]), body[1])
			}
		}
	})

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go processRequest(conn)
	}
}

func processRequest(conn net.Conn) {
	buffer := make([]byte, 1024)
	conn.Read(buffer)
	r := request(strings.TrimRight(string(buffer), "\x00"))
	handleRequest(r, conn)
}
