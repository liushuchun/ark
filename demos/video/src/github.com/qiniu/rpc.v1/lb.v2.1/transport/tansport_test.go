package transport

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWarpRequest(t *testing.T) {
	old := MaxBodyLength
	defer func() {
		MaxBodyLength = old
	}()
	MaxBodyLength = 2
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	_, err = warpRequest(req)
	assert.Equal(t, ErrTooLargeBody, err)

	body = "a=2"
	req, err = http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = -1
	_, err = warpRequest(req)
	assert.Equal(t, ErrTooLargeBody, err)
}
