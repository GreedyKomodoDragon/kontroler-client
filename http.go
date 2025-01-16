package kontrolerclient

import (
	"context"
	"net/http"
	"sync"
)

type contextKey string

const retryKey contextKey = "retry"

type cookieTransport struct {
	cookie    *http.Cookie
	transport http.RoundTripper
	client    *client
	mu        sync.Mutex
}

// RoundTrip executes a single HTTP transaction and retries login if unauthorized
func (t *cookieTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add the cookie to the request - could change if the cookie is updated
	t.mu.Lock()
	req.AddCookie(t.cookie)
	t.mu.Unlock()

	// Perform the HTTP request
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		t.mu.Lock()
		defer t.mu.Unlock()

		// Retry login once by checking the context for the retry key
		if req.Context().Value(retryKey) == nil {
			// Set the retry key in the context to avoid infinite loops
			ctx := context.WithValue(req.Context(), retryKey, true)
			req = req.WithContext(ctx)

			if err := t.client.login(t.client.username, t.client.password); err != nil {
				return nil, err
			}

			req.AddCookie(t.cookie)
			resp, err = t.transport.RoundTrip(req)
		}
	}

	return resp, err
}
