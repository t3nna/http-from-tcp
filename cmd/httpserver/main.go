package main

import (
	"fmt"
	"github.com/t3nna/http-from-tcp/internal/request"
	"github.com/t3nna/http-from-tcp/internal/response"
	"github.com/t3nna/http-from-tcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func respond400() []byte {
	return []byte(`
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`)
}
func respond500() []byte {
	return []byte(`
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`)
}
func respond200() []byte {
	return []byte(`
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`)
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		h.Replace("content-type", "text/html")
		body := respond200()

		if req.RequestLine.RequestTarget == "/yourproblem" {
			w.WriteStatusLine(response.StatusBarRequest)

			body = respond400()
			h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteHeaders(h)
			w.WriteBody(body)
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			w.WriteStatusLine(response.StatusInternalServerError)

			body = respond500()
			h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteHeaders(h)
			w.WriteBody(body)

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/stream") {
			target := req.RequestLine.RequestTarget
			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				body = respond500()
			} else {
				w.WriteStatusLine(response.StatusOK)
				h.Set("transfer-encoding", "chunked")
				h.Delete("content-length")
				h.Replace("content-type", "text/plain")
				w.WriteHeaders(h)

				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}

					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n\r\n"))

				return
			}

		}

		w.WriteStatusLine(response.StatusOK)

		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeaders(h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting s: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
