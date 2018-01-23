package jsonlog

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify.v1/require"
)

type simpleResponseWriter struct {
	http.ResponseWriter
}

func TestGetExtraWriter(t *testing.T) {

	r := &responseWriter{}

	var w http.ResponseWriter

	w = &simpleResponseWriter{r}
	_, ok := getExtraWriter(w)
	require.True(t, ok)

	w = &simpleResponseWriter{&simpleResponseWriter{r}}
	_, ok = getExtraWriter(w)
	require.True(t, ok)
}
