package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer f.Close()
		defer close(ch)
		fullLine := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}
			data = data[:n]

			if index := bytes.IndexByte(data, '\n'); index != -1 {
				fullLine += string(data[:index])
				ch <- fullLine
				fullLine = ""
				data = data[index+1:]
			}
			fullLine += string(data)

		}
		if fullLine != "" {
			ch <- fullLine
		}

	}()
	return ch
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}

		fmt.Println("Accepted connection from", conn.RemoteAddr())

		for line := range getLinesChannel(conn) {

			fmt.Printf("read: %s \n", line)
		}
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}
