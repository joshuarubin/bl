package client

import "net/http"

type transport struct {
	Transport http.RoundTripper
	Calls     int
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.Calls++
	return t.Transport.RoundTrip(r)
}
