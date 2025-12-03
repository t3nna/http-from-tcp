package headers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid header
	headers := NewHeaders()
	data := []byte("hOSt: localhost:42069\r\nFoo:bar\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	headerName, _ := headers.Get("HoST")
	assert.Equal(t, "localhost:42069", headerName)
	headerName, _ = headers.Get("Foo")
	assert.Equal(t, "bar", headerName)
	headerName, _ = headers.Get("Missing Key")
	assert.Equal(t, "", headerName)
	assert.Equal(t, 34, n)
	assert.True(t, done)

	// Test: Valid with same headers
	headers = NewHeaders()
	data = []byte("host: localhost:42069\r\nhost: bar\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	headerName, _ = headers.Get("host")
	assert.Equal(t, "localhost:42069,bar", headerName)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid char in header name
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
