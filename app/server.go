package main

import (
	"fmt"
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
