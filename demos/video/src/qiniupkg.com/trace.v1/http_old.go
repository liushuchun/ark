// +build !go1.8

package trace

import "net/http"

func newClientEvent(t *Recorder, r *http.Request) *ClientEvent {
	if r.Body != nil {
		r.Body = ReadCloserWithTrace(t, r.Body, "req.body.send")
	}
	return &ClientEvent{
		t:       t,
		Request: requestInfo(r),
	}
}
