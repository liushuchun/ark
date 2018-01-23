package httputil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify.v1/require"
)

// ---------------------------------------------------------------------------

type oneTransport struct {
	Transport http.RoundTripper
}

func (t oneTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	return t.Transport.RoundTrip(req)
}

// ---------------------------------------------------------------------------

type twoTransport struct {
	a   int
	b   int
	One http.RoundTripper
}

func (t *twoTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	return t.One.RoundTrip(req)
}

// ---------------------------------------------------------------------------

type threeTransport struct {
	Two *twoTransport
}

func (t *threeTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	return t.Two.RoundTrip(req)
}

// ---------------------------------------------------------------------------

func TestGetRequestCanceler(t *testing.T) {

	zero := http.DefaultTransport
	one := oneTransport{zero}
	two := &twoTransport{1, 2, one}
	three := &threeTransport{two}

	if _, ok := GetRequestCanceler(zero); !ok {
		t.Fatal("GetRequestCanceler(zero) failed")
	}

	if _, ok := GetRequestCanceler(one); !ok {
		t.Fatal("GetRequestCanceler(one) failed")
	}

	if _, ok := GetRequestCanceler(two); !ok {
		t.Fatal("GetRequestCanceler(two) failed")
	}

	if _, ok := GetRequestCanceler(three); !ok {
		t.Fatal("GetRequestCanceler(three) failed")
	}
}

// ---------------------------------------------------------------------------

type flushedResponseWriter struct {
	http.ResponseWriter
}

func (w *flushedResponseWriter) Flush() {
}

type wrappedResponseWriter struct {
	http.ResponseWriter
}

func TestFlusher(t *testing.T) {

	r := &flushedResponseWriter{}

	var w http.ResponseWriter
	w = r

	f, ok := w.(http.Flusher)
	require.True(t, f != nil && ok)
	f, ok = Flusher(w)
	require.True(t, f != nil && ok)

	w = &wrappedResponseWriter{r}
	f, ok = w.(http.Flusher)
	require.False(t, ok)
	f, ok = Flusher(w)
	require.True(t, f != nil && ok)

	w = &wrappedResponseWriter{&wrappedResponseWriter{r}}
	f, ok = w.(http.Flusher)
	require.False(t, ok)
	f, ok = Flusher(w)
	require.True(t, f != nil && ok)
}
