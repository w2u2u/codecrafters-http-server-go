package main

import (
	"fmt"
	"net"
	"slices"
	"strings"
)

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
			accept_encoding := strings.SplitN(line, " ", 2)[1]
			encodings := strings.Split(accept_encoding, ", ")

			if slices.Contains(encodings, "gzip") {
				request.acceptEncoding = "gzip"
			}

			fmt.Println("request", request)
		}
	}

	if bufferLines[len(bufferLines)-1] != "" {
		request.body = strings.TrimSpace(bufferLines[len(bufferLines)-1])
	}

	return request, nil
}
