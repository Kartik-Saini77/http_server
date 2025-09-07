package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
	
	// Test: Duplicate header key
	headers = NewHeaders()
	lines := []string{
		"Set-Person: lane-loves-go\r\n",
		"Set-Person: prime-loves-zig\r\n",
		"Set-Person: tj-loves-ocaml\r\n",
	}
	for _, line := range lines {
		n, done, err := headers.Parse([]byte(line))
		require.NoError(t, err)
		require.False(t, done)
		require.Equal(t, len(line), n)
	}
	expected := "lane-loves-go, prime-loves-zig, tj-loves-ocaml"
	assert.Equal(t, expected, headers["set-person"])
}
