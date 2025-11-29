package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	f, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatalf("cant open the file: %v", err)
	}
	defer f.Close()

	var line []byte
	// 1. Move allocation outside the loop
	buf := make([]byte, 8)

	for {
		n, err := f.Read(buf)
		// Handle EOF specifically before checking other errors
		if n > 0 {
			// Process the data we just read
			data := buf[:n]

			// 2. Loop to find ALL newlines in this chunk
			for {
				idx := bytes.IndexByte(data, '\n')
				if idx == -1 {
					// No more newlines in this chunk, append remainder and break inner loop
					line = append(line, data...)
					break
				}

				// Append up to the newline
				line = append(line, data[:idx]...)

				// We have a full line now
				fmt.Printf("read: %s \n", line)

				// Reset line buffer
				line = line[:0]

				// Advance data slice past the newline we just processed
				data = data[idx+1:]
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading: %v", err)
		}
	}

	// 3. Handle any remaining data after EOF (no ending newline in file)
	if len(line) > 0 {
		fmt.Printf("read: %s \n", line)
	}
}
