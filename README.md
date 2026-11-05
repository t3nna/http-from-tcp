# HTTP from TCP

This Go learning project explores how an HTTP server works by building parts of HTTP/1.1 on top of TCP sockets. It parses requests and writes responses using the project's own packages, rather than using Go's `net/http` server. The streaming example uses `net/http` as a client to fetch an upstream response.

## How it works

1. The server listens for TCP connections on port `42069` and handles each connection in a goroutine.
2. The request parser reads bytes from the connection incrementally. It parses the request line, headers, and an optional body specified by `Content-Length`.
3. The server passes the parsed request and a response writer to a handler.
4. The handler writes an HTTP status line, headers, and body. The connection is then closed.

The implementation is organized into four packages under `internal/`:

- `request` parses HTTP/1.1 request lines, headers, and bodies, including data split across multiple reads.
- `headers` validates and stores header fields with case-insensitive names; repeated fields are combined.
- `response` writes status lines, headers, and response bodies.
- `server` accepts TCP connections and connects the parser and response writer to a handler.

## Run the HTTP server

The module specifies Go **1.25.3**. From the repository root, run:

```sh
go run ./cmd/httpserver
```

In another terminal, try the routes:

```sh
curl -i http://localhost:42069/
curl -i http://localhost:42069/yourproblem
curl -i http://localhost:42069/myproblem
curl -i --raw http://localhost:42069/httpbin/stream/3
curl -i http://localhost:42069/video
```

| Route | Behavior |
| --- | --- |
| `/yourproblem` | Returns an HTML response with status 400. |
| `/myproblem` | Returns an HTML response with status 500. |
| `/httpbin/stream...` | Fetches the corresponding stream from `https://httpbin.org/` and sends it using chunked transfer encoding, with SHA-256 and length trailers. |
| `/video` | Reads `assets/vim.mp4` relative to the working directory and sends it as `video/mp4`. |
| Other paths | Return an HTML success response with status 200. |

The streaming route requires network access to httpbin.org. The video file is not included in the repository; provide `assets/vim.mp4` from the repository root to try that route. The current error handling for these examples is incomplete, so an upstream or file error may not produce the expected error response.

## Other exercises

Run these from the repository root:

```sh
go run ./cmd/tcplistener
go run ./cmd/udpsender
go run ./readFromFile.go
```

The TCP listener accepts connections on port `42069` and prints parsed request lines and headers. Stop the HTTP server before starting it because both use the same TCP port. The UDP sender sends lines entered on standard input to `localhost:42069`; it needs a separate UDP receiver. The TCP listener cannot receive UDP messages. The file-reading exercise reads `messages.txt` in small chunks and prints its lines.

## Tests

```sh
go test ./...
```

The tests cover header parsing and request parsing across different read sizes, including request bodies and malformed input.

## Current limitations

This is an educational, incomplete HTTP implementation, not a production server. The request parser supports HTTP/1.1 and `Content-Length` bodies, but does not implement chunked request bodies or persistent connections. Request validation and error handling are limited, and the server's shutdown method does not currently close its listener. The HTTP server closes each connection after handling one request.
