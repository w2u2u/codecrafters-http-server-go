package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

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

	handle_connection(conn)
}

func handle_connection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)

	conn.Read(buffer)

	req, err := NewRequest(buffer)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
	}

	fmt.Printf("Request: %+v", req)

	switch true {
	case req.method == "GET" && req.path == "/":
		handle_index(conn)
	default:
		response_not_found(conn)
	}
}

type Request struct {
	method     string
	path       string
	user_agent string
	body       string
}

func NewRequest(buffer []byte) (Request, error) {
	buffer_lines := strings.Split(string(buffer), "\r\n")

	if len(buffer_lines) < 2 {
		return Request{}, fmt.Errorf("buffer: invalid number of lines: %d", len(buffer_lines))
	}

	first_lines := strings.Split(buffer_lines[0], " ")

	if len(first_lines) != 3 {
		return Request{}, fmt.Errorf("buffer: invalid http header: %v", first_lines)
	}

	// fmt.Printf("buffer_lines: %v\n", buffer_lines)
	// fmt.Printf("first_lines: %v\n", first_lines)

	method := first_lines[0]
	path := first_lines[1]
	user_agent := strings.Split(buffer_lines[2], " ")[1]
	body := ""

	return Request{method, path, user_agent, body}, nil
}

func handle_index(conn net.Conn) {
	response_ok(conn, "", "")
}

func response_ok(conn net.Conn, content_type string, content string) {
	if len(content_type) == 0 || len(content) == 0 {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", content_type, len(content), content)

		conn.Write([]byte(response))
	}
}

func response_not_found(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}
