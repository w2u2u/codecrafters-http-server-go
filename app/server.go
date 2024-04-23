package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	args := os.Args
	directory := get_arg(args, "--directory")

	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handle_connection(conn, directory)
	}
}

func get_arg(args []string, name string) string {
	for i, arg := range args {
		if arg == name && len(args) > i {
			return args[i+1]
		}
	}

	return ""
}

func handle_connection(conn net.Conn, directory string) {
	defer conn.Close()
	defer fmt.Println("Closing connection")

	req, err := NewRequest(conn)
	if err != nil {
		fmt.Println(err)
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	switch true {
	case req.method == "GET" && req.path == "/":
		handle_index(conn)
	case req.method == "GET" && strings.HasPrefix(req.path, "/echo"):
		handle_echo(conn, req)
	case req.method == "GET" && strings.HasPrefix(req.path, "/user-agent"):
		handle_user_agent(conn, req)
	case req.method == "GET" && strings.HasPrefix(req.path, "/files"):
		handle_read_file(conn, req, directory)
	case req.method == "POST" && strings.HasPrefix(req.path, "/files"):
		handle_write_file(conn, req, directory)
	default:
		not_found(conn)
	}
}

type Request struct {
	method     string
	path       string
	user_agent string
	body       string
}

func NewRequest(conn net.Conn) (Request, error) {
	buffer := make([]byte, 1024)

	length, _ := conn.Read(buffer)

	buffer_lines := strings.Split(string(buffer[:length]), "\r\n")

	if len(buffer_lines) == 0 {
		return Request{}, fmt.Errorf("buffer: invalid number of lines: %d", len(buffer_lines))
	}

	first_lines := strings.Split(buffer_lines[0], " ")

	if len(first_lines) != 3 {
		return Request{}, fmt.Errorf("buffer: invalid http header: %v", first_lines)
	}

	method := first_lines[0]
	path := first_lines[1]
	user_agent := ""
	body := ""

	if len(buffer_lines) > 2 && buffer_lines[2] != "" {
		user_agent = strings.Split(buffer_lines[2], " ")[1]
	}

	if buffer_lines[len(buffer_lines)-1] != "" {
		body = strings.TrimSpace(buffer_lines[len(buffer_lines)-1])
	}

	return Request{method, path, user_agent, body}, nil
}

func handle_index(conn net.Conn) {
	ok(conn, "", "")
}

func handle_echo(conn net.Conn, req Request) {
	path := strings.TrimLeft(req.path, "/echo/")

	ok(conn, "text/plain", path)
}

func handle_user_agent(conn net.Conn, req Request) {
	ok(conn, "text/plain", req.user_agent)
}

func handle_read_file(conn net.Conn, req Request, directory string) {
	file_name := strings.TrimLeft(req.path, "/files")
	file_path := fmt.Sprintf("%s%s", directory, file_name)

	content, err := os.ReadFile(file_path)
	if err != nil {
		log.Println("[ERR]", err)
		not_found(conn)
		return
	}

	ok(conn, "application/octet-stream", string(content))
}

func handle_write_file(conn net.Conn, req Request, directory string) {
	file_name := strings.TrimLeft(req.path, "/files")
	file_path := fmt.Sprintf("%s%s", directory, file_name)

	err := os.WriteFile(file_path, []byte(req.body), 0644)
	if err != nil {
		log.Println("[ERR]", err)
		not_found(conn)
		return
	}

	created(conn)
}

func ok(conn net.Conn, content_type string, content string) {
	if len(content_type) == 0 || len(content) == 0 {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", content_type, len(content), content)

		conn.Write([]byte(response))
	}
}

func created(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 201 CREATED\r\n\r\n"))
}

func not_found(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}
