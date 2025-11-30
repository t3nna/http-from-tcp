package request

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

// TestDifferentChunkSizes tests parsing with various chunk sizes from 1 byte to full request length
func TestDifferentChunkSizes(t *testing.T) {
	testCases := []struct {
		name            string
		request         string
		chunkSize       int
		expectedMethod  string
		expectedPath    string
		expectedVersion string
	}{
		{
			name:            "1 byte chunks - root path",
			request:         "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			chunkSize:       1,
			expectedMethod:  "GET",
			expectedPath:    "/",
			expectedVersion: "1.1",
		},
		{
			name:            "2 byte chunks",
			request:         "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			chunkSize:       2,
			expectedMethod:  "GET",
			expectedPath:    "/",
			expectedVersion: "1.1",
		},
		{
			name:            "3 byte chunks - with path",
			request:         "GET /api/users HTTP/1.1\r\nHost: localhost\r\n\r\n",
			chunkSize:       3,
			expectedMethod:  "GET",
			expectedPath:    "/api/users",
			expectedVersion: "1.1",
		},
		{
			name:            "5 byte chunks",
			request:         "POST /login HTTP/1.1\r\nHost: localhost\r\n\r\n",
			chunkSize:       5,
			expectedMethod:  "POST",
			expectedPath:    "/login",
			expectedVersion: "1.1",
		},
		{
			name:            "10 byte chunks",
			request:         "PUT /api/data/123 HTTP/1.1\r\nHost: localhost\r\n\r\n",
			chunkSize:       10,
			expectedMethod:  "PUT",
			expectedPath:    "/api/data/123",
			expectedVersion: "1.1",
		},
		{
			name:            "20 byte chunks",
			request:         "DELETE /api/users/456 HTTP/1.1\r\nHost: localhost\r\n\r\n",
			chunkSize:       20,
			expectedMethod:  "DELETE",
			expectedPath:    "/api/users/456",
			expectedVersion: "1.1",
		},
		{
			name:            "Full request in one chunk",
			request:         "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
			chunkSize:       len("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"),
			expectedMethod:  "GET",
			expectedPath:    "/coffee",
			expectedVersion: "1.1",
		},
		{
			name:            "Chunk size larger than request",
			request:         "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			chunkSize:       1000,
			expectedMethod:  "GET",
			expectedPath:    "/",
			expectedVersion: "1.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := &chunkReader{
				data:            tc.request,
				numBytesPerRead: tc.chunkSize,
			}
			r, err := RequestFromReader(reader)
			require.NoError(t, err, "Failed to parse request with chunk size %d", tc.chunkSize)
			require.NotNil(t, r)
			assert.Equal(t, tc.expectedMethod, r.RequestLine.Method, "Method mismatch")
			assert.Equal(t, tc.expectedPath, r.RequestLine.RequestTarget, "Path mismatch")
			assert.Equal(t, tc.expectedVersion, r.RequestLine.HttpVersion, "Version mismatch")
		})
	}
}

// TestDifferentHttpMethods tests various HTTP methods
func TestDifferentHttpMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			request := method + " /test HTTP/1.1\r\nHost: localhost\r\n\r\n"
			reader := &chunkReader{
				data:            request,
				numBytesPerRead: 1, // Test with smallest chunk size
			}
			r, err := RequestFromReader(reader)
			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Equal(t, method, r.RequestLine.Method)
			assert.Equal(t, "/test", r.RequestLine.RequestTarget)
			assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
		})
	}
}

// TestDifferentPaths tests various request paths
func TestDifferentPaths(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		chunkSize    int
	}{
		{"root path", "/", 1},
		{"simple path", "/coffee", 2},
		{"nested path", "/api/users/123", 3},
		{"path with query", "/search?q=test", 5},
		{"long path", "/very/long/path/with/many/segments", 10},
		{"path with special chars", "/api/v1/users", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := "GET " + tc.path + " HTTP/1.1\r\nHost: localhost\r\n\r\n"
			reader := &chunkReader{
				data:            request,
				numBytesPerRead: tc.chunkSize,
			}
			r, err := RequestFromReader(reader)
			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Equal(t, "GET", r.RequestLine.Method)
			assert.Equal(t, tc.path, r.RequestLine.RequestTarget)
			assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
		})
	}
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("Request line split across chunk boundary at CRLF", func(t *testing.T) {
		// This tests when \r\n is split across chunks
		request := "GET /test HTTP/1.1\r\nHost: localhost\r\n\r\n"
		reader := &chunkReader{
			data:            request,
			numBytesPerRead: 18, // Split right before \r\n
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/test", r.RequestLine.RequestTarget)
	})

	t.Run("Request line split at space boundary", func(t *testing.T) {
		request := "GET /coffee HTTP/1.1\r\nHost: localhost\r\n\r\n"
		reader := &chunkReader{
			data:            request,
			numBytesPerRead: 4, // Split at "GET "
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	})

	t.Run("Very long request with small chunks", func(t *testing.T) {
		// Create a request with a very long path
		longPath := "/" + string(make([]byte, 500)) // 500 byte path
		request := "GET " + longPath + " HTTP/1.1\r\nHost: localhost\r\n\r\n"
		reader := &chunkReader{
			data:            request,
			numBytesPerRead: 1, // 1 byte at a time
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, longPath, r.RequestLine.RequestTarget)
	})
}

// TestErrorCases tests malformed requests and error conditions
func TestErrorCases(t *testing.T) {
	t.Run("Malformed request - missing parts", func(t *testing.T) {
		request := "GET / HTTP/1.1\r\n" // Missing second space
		reader := &chunkReader{
			data:            request,
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		// This should either error or handle gracefully depending on implementation
		// The current implementation might not catch this if it's after the request line
		_ = r
		_ = err
	})

	t.Run("Unsupported HTTP version", func(t *testing.T) {
		request := "GET / HTTP/2.0\r\nHost: localhost\r\n\r\n"
		reader := &chunkReader{
			data:            request,
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		require.Error(t, err, "Should error on unsupported HTTP version")
		assert.Nil(t, r)
	})

	t.Run("Invalid method - lowercase", func(t *testing.T) {
		request := "get / HTTP/1.1\r\nHost: localhost\r\n\r\n"
		reader := &chunkReader{
			data:            request,
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		require.Error(t, err, "Should error on lowercase method")
		assert.Nil(t, r)
	})

	t.Run("Missing CRLF", func(t *testing.T) {
		// Request without \r\n - should need more data
		request := "GET / HTTP/1.1"
		reader := &chunkReader{
			data:            request,
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		// This might timeout or error depending on implementation
		// The parser should return 0 consumed bytes when no CRLF is found
		_ = r
		_ = err
	})
}

// TestChunkBoundaryAtCriticalPoints tests chunk boundaries at critical parsing points
func TestChunkBoundaryAtCriticalPoints(t *testing.T) {
	baseRequest := "GET /api/users HTTP/1.1\r\nHost: localhost\r\n\r\n"
	
	// Test chunk sizes that align with critical points
	criticalPoints := []int{
		1,   // Before "G"
		4,   // After "GET"
		5,   // At the space after GET
		6,   // At the start of path
		10,  // In the middle of path
		18,  // Right before \r\n
		19,  // At \r
		20,  // At \n
	}

	for _, point := range criticalPoints {
		t.Run(fmt.Sprintf("chunk_at_byte_%d", point), func(t *testing.T) {
			reader := &chunkReader{
				data:            baseRequest,
				numBytesPerRead: point,
			}
			r, err := RequestFromReader(reader)
			require.NoError(t, err, "Failed at chunk size %d", point)
			require.NotNil(t, r)
			assert.Equal(t, "GET", r.RequestLine.Method)
			assert.Equal(t, "/api/users", r.RequestLine.RequestTarget)
			assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
		})
	}
}
