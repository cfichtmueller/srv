// The proxy package is experimental. The API will most likely change.
package proxy

import (
	"io"
	"net/http"
)

// Proxy represents a proxy that can forward requests to an upstream server.
type Proxy struct {
	UpstreamURL string
}

// NewProxy creates a new Proxy with the given upstream URL.
func NewProxy(upstreamURL string) *Proxy {
	return &Proxy{
		UpstreamURL: upstreamURL,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, p.UpstreamURL+r.RequestURI, r.Body)
	if err != nil {
		http.Error(w, "Failed to create upstream request", http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request
	for name, values := range r.Header {
		for _, v := range values {
			req.Header.Add(name, v)
		}
	}

	clientIp := r.Header.Get("X-Forwarded-For")
	if clientIp == "" {
		clientIp = r.RemoteAddr
	}
	req.Header.Set("X-Forwarded-For", clientIp)
	req.Header.Set("X-Real-IP", clientIp)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Upstream request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy upstream response headers and status
	for name, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(name, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Copy upstream response body
	_, _ = io.Copy(w, resp.Body)
}
