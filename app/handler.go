package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

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
	paths := strings.Split(req.path, "/")
	fileName := paths[2]
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
	paths := strings.Split(req.path, "/")
	fileName := paths[2]
	filePath := fmt.Sprintf("%s%s", directory, fileName)

	err := os.WriteFile(filePath, []byte(req.body), 0644)
	if err != nil {
		log.Println("[ERR]", err)
		notFound(conn)
		return
	}

	created(conn)
}
