package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
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
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatalf("unable to read a file %v", err)
	}

	ch := getLinesChannel(file)
	for val := range ch {
		fmt.Printf("read: %s \n", val)
	}

	// leftover
	//if len(line) != 0 {
	//	fmt.Printf("read: %s \n", line)
	//}

}
