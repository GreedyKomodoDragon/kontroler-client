package kontrolerclient

import "net/http"

type cookieTransport struct {
	cookie    *http.Cookie
	transport http.RoundTripper
}

func (t *cookieTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.AddCookie(t.cookie)
	return t.transport.RoundTrip(req)
}
