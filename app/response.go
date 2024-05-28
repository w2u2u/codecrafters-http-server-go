package main

import (
	"fmt"
	"net"
)

func ok(conn net.Conn, contentType string, acceptEncoding string, content string) {
	if len(contentType) == 0 || len(content) == 0 {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		response := "HTTP/1.1 200 OK"
		response = fmt.Sprintf("%s\r\nContent-Type: %s", response, contentType)
		response = fmt.Sprintf("%s\r\nContent-Length: %d", response, len(content))

		if len(acceptEncoding) > 0 {
			response = fmt.Sprintf("%s\r\nContent-Encoding: %s", response, acceptEncoding)
		}

		response = fmt.Sprintf("%s\r\n\r\n%s", response, content)

		conn.Write([]byte(response))
	}
}

func created(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
}

func notFound(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}
