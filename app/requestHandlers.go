package main

import (
	"fmt"
	"net"
	"strings"
)

type handlers map[string]func(request, net.Conn)

var getHandlers = make(handlers)
var postHandlers = make(handlers)

func registerGetHandler(target string, handler func(request, net.Conn)) {
	getHandlers[target] = handler
}

func registerPostHandler(target string, handler func(request, net.Conn)) {
	postHandlers[target] = handler
}

func handleRequest(r request, conn net.Conn) {
	verb, error := r.getVerb()

	if error != nil {
		fmt.Fprintf(conn, "HTTP/1.1 405 Method Not Allowed\r\n\r\n%s", error)
	}

	hs := make(handlers)

	switch verb {
	case get:
		hs = getHandlers
	case post:
		hs = postHandlers
	}

	target, error := r.getTarget()

	if error != nil {
		fmt.Fprintf(conn, "HTTP/1.1 400 Bad Request\r\n\r\n%s", error)
		return
	}

	handler := hs[strings.Split(target, "/")[1]]

	if handler == nil {
		fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
		return
	}

	handler(r, conn)
}
