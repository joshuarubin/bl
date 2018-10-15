package client

import (
	"net/http"
	"sync/atomic"
)

type transport struct {
	Transport http.RoundTripper
	Calls     int32
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	atomic.AddInt32(&t.Calls, 1)
	return t.Transport.RoundTrip(r)
}
