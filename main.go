package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatalf("unable to read a file %v", err)
	}
	defer file.Close()

	line := ""

	for {
		data := make([]byte, 8)
		n, err := file.Read(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
		data = data[:n]

		if i := bytes.IndexByte(data, '\n'); i != -1 {
			line += string(data[:i])

			data = data[i+1:]
			fmt.Printf("read: %s \n", line)

			line = ""
		}
		line += string(data)

	}

	// leftover
	if len(line) != 0 {
		fmt.Printf("read: %s \n", line)
	}

}
