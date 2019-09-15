package kproxy

import (
	"io"
	"net/http"
)

func (proxy *ProxyServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	logRequest("HTTP Request", r)

	// Error against a non-proxy request
	if !r.URL.IsAbs() {
		httpError(w, "Non-Proxy request is not supported", http.StatusBadRequest, nil)
		return
	}

	// Create a new request to send to the remote host
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		httpError(w, "Failed to create a new request", http.StatusInternalServerError, err)
		return
	}
	copyHeadersForProxy(req.Header, r.Header)

	// Call OnBeforeRequest handler
	if proxy.OnBeforeRequest != nil {
		req = proxy.OnBeforeRequest(req)
	}

	// Send a request with OnSendRequest handler or the original http client.
	var resp *http.Response
	if proxy.OnSendRequest != nil {
		resp, err = proxy.OnSendRequest(req)
	} else {
		resp, err = proxy.httpClient.Do(req)
	}
	if err != nil {
		httpError(w, "Failed to send a request to server", http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	// Call OnBeforeResponse handler
	if proxy.OnBeforeResponse != nil {
		resp = proxy.OnBeforeResponse(resp)
	}
	logResponse("HTTP Response", resp)

	// Return HTTP status code and headers to the client
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	// Return the body
	if _, err := io.Copy(w, resp.Body); err != nil {
		klog("Failed to copy data : %v", err)
	}
}
