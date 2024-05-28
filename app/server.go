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
	directory := getArg(args, "--directory")

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

		go handleConnection(conn, directory)
	}
}

func getArg(args []string, name string) string {
	for i, arg := range args {
		if arg == name && len(args) > i {
			return args[i+1]
		}
	}

	return ""
}

func handleConnection(conn net.Conn, directory string) {
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
		handleIndex(conn)
	case req.method == "GET" && strings.HasPrefix(req.path, "/echo"):
		handleEcho(conn, req)
	case req.method == "GET" && strings.HasPrefix(req.path, "/user-agent"):
		handleUserAgent(conn, req)
	case req.method == "GET" && strings.HasPrefix(req.path, "/files"):
		handleReadFile(conn, req, directory)
	case req.method == "POST" && strings.HasPrefix(req.path, "/files"):
		handleWriteFile(conn, req, directory)
	default:
		notFound(conn)
	}
}

type Request struct {
	method         string
	path           string
	userAgent      string
	body           string
	acceptEncoding string
}

func NewRequest(conn net.Conn) (Request, error) {
	buffer := make([]byte, 1024)

	length, _ := conn.Read(buffer)

	bufferLines := strings.Split(string(buffer[:length]), "\r\n")

	if len(bufferLines) == 0 {
		return Request{}, fmt.Errorf("buffer: invalid number of lines: %d", len(bufferLines))
	}

	firstLines := strings.Split(bufferLines[0], " ")

	if len(firstLines) != 3 {
		return Request{}, fmt.Errorf("buffer: invalid http header: %v", firstLines)
	}

	request := Request{
		method: firstLines[0],
		path:   firstLines[1],
	}

	for _, line := range bufferLines {
		switch true {
		case strings.HasPrefix(line, "User-Agent"):
			request.userAgent = strings.Split(line, " ")[1]
		case strings.HasPrefix(line, "Accept-Encoding"):
			accept_encoding := strings.Split(line, " ")[1]
			if accept_encoding == "gzip" {
				request.acceptEncoding = strings.Split(line, " ")[1]
			}
		}
	}

	if bufferLines[len(bufferLines)-1] != "" {
		request.body = strings.TrimSpace(bufferLines[len(bufferLines)-1])
	}

	return request, nil
}

func handleIndex(conn net.Conn) {
	ok(conn, "", "", "")
}

func handleEcho(conn net.Conn, req Request) {
	paths := strings.Split(req.path, "/")

	ok(conn, "text/plain", req.acceptEncoding, paths[2])
}

func handleUserAgent(conn net.Conn, req Request) {
	ok(conn, "text/plain", req.acceptEncoding, req.userAgent)
}

func handleReadFile(conn net.Conn, req Request, directory string) {
	fileName := strings.TrimLeft(req.path, "/files")
	filePath := fmt.Sprintf("%s%s", directory, fileName)

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("[ERR]", err)
		notFound(conn)
		return
	}

	ok(conn, "application/octet-stream", req.acceptEncoding, string(content))
}

func handleWriteFile(conn net.Conn, req Request, directory string) {
	fileName := strings.TrimLeft(req.path, "/files")
	filePath := fmt.Sprintf("%s%s", directory, fileName)

	err := os.WriteFile(filePath, []byte(req.body), 0644)
	if err != nil {
		log.Println("[ERR]", err)
		notFound(conn)
		return
	}

	created(conn)
}

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
