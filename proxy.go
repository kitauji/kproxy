package kproxy

import (
	"crypto/tls"
	"net/http"
)

// ProxyServer is a proxy object that implements ServeHTTP() function.
type ProxyServer struct {
	mitmMode   bool
	httpClient *http.Client

	// This map contains tls.Certificate that will be generated
	// in MITM mode. The key is hostname that a client connected to.
	certificates map[string]*tls.Certificate

	// OnBeforeRequest is a hook point that can be used
	// to customize a request before sending it to the remote host.
	OnBeforeRequest func(r *http.Request) *http.Request

	// OnSendRequest is a hook point that can be used
	// to send a request instead of kproxy.
	// It must return a *http.Response.
	OnSendRequest func(r *http.Request) (*http.Response, error)

	// OnBeforeResponse is a hook point that can be used
	// to customize a response before returning it to the client.
	OnBeforeResponse func(resp *http.Response) *http.Response
}

// NewProxyServer create a new ProxyServer instance.
func NewProxyServer(mitmMode bool) *ProxyServer {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}

	return &ProxyServer{
		mitmMode:     mitmMode,
		httpClient:   client,
		certificates: make(map[string]*tls.Certificate),
	}
}

func (proxy *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		// HTTPS CONNECT
		proxy.handleConnect(w, r)
	} else {
		// HTTP
		proxy.handleHTTP(w, r)
	}
}
