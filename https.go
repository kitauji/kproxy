package kproxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"

	kcert "github.com/kitauji/kutils/certificate"
)

func hijackConnection(w http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("Failed to get Hijacker interface")
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, errors.New("Failed to hijack TCP connection : " + err.Error())
	}

	return conn, nil
}

func (proxy *ProxyServer) getTLSConfig(hostname string) (*tls.Config, error) {

	cert, ok := proxy.certificates[hostname]
	if !ok {
		c, err := kcert.CreateCertificate([]string{hostname}, 365)
		if err != nil {
			return nil, err
		}
		cert = c
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		InsecureSkipVerify: true,
	}

	return tlsConfig, nil
}

func (proxy *ProxyServer) handleHTTPSWithMITM(w http.ResponseWriter, r *http.Request) {
	// Hijack TCP connection underlaying the client's request to transfer HTTPS stream
	clientConn, err := hijackConnection(w)
	if err != nil {
		// We might not be able to send an HTTP error to the client...
		httpError(w, "Failed to hijack a client connection", http.StatusInternalServerError, err)
		return
	}

	clientConn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	hostname := parseHostname(r.Host)
	tlsConfig, err := proxy.getTLSConfig(hostname)
	if err != nil {
		clientConn.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n"))
		clientConn.Close()
		return
	}

	go func() {
		clientTLSConn := tls.Server(clientConn, tlsConfig)
		defer clientTLSConn.Close()

		if err := clientTLSConn.Handshake(); err != nil {
			klog("TLS Handshake error : %v", err)
			return
		}

		clientTLSReader := bufio.NewReader(clientTLSConn)

		for {
			req, err := http.ReadRequest(clientTLSReader)
			if err != nil {
				if err != io.EOF {
					klog("TLS connection error : %v", err)
				}
				return
			}

			logRequest("HTTPS request in MITM mode", req)

			// Create a new request to send to the remote host
			host := parseHostname(req.Host)
			req.URL.Scheme = "https"
			req.URL.Host = host
			URL := req.URL.String()

			newReq, err := http.NewRequest(req.Method, URL, req.Body)
			if err != nil {
				klog("Failed to create an HTTPS request : %v", err)
				return
			}
			copyHeadersForProxy(newReq.Header, req.Header)

			// Call OnBeforeRequest handler
			if proxy.OnBeforeRequest != nil {
				newReq = proxy.OnBeforeRequest(newReq)
			}

			logRequest("New HTTPS request in MITM mode", newReq)

			// Send a request with OnSendRequest handler or the original http client.
			var resp *http.Response
			if proxy.OnSendRequest != nil {
				resp, err = proxy.OnSendRequest(newReq)
			} else {
				resp, err = proxy.httpClient.Do(newReq)
			}
			if err != nil {
				klog("Failed to send an HTTPS request : %v", err)
				return
			}
			defer resp.Body.Close()

			// Since resp.ContentLength is always -1, we have to return the response
			// with "chunked" mode, so that the client can know the end of the response.
			resp.TransferEncoding = []string{"chunked"}
			resp.Header.Set("Connection", "close")

			// Call OnBeforeResponse handler
			if proxy.OnBeforeResponse != nil {
				resp = proxy.OnBeforeResponse(resp)
			}

			logResponse("HTTPS response", resp)

			// Write the response to client's TLS connection
			if err := resp.Write(clientTLSConn); err != nil {
				klog("Failed to send a HTTPS response : %v", err)
				return
			}
		}
	}()
}

// handleHTTPS handles CONNECT method and subsequent requests with MITM disabled.
func (proxy *ProxyServer) handleHTTPS(w http.ResponseWriter, r *http.Request) {

	// Hijack TCP connection underlaying the client's request to transfer HTTPS stream
	clientConn, err := hijackConnection(w)
	if err != nil {
		// We might not be able to send an HTTP error to the client...
		httpError(w, "Failed to hijack a client connection", http.StatusInternalServerError, err)
		return
	}

	// Connect to the remote server
	serverConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		klog("Failed to connecto the remote server : %v", err)
		clientConn.Write([]byte("HTTP/1.0 502 Bad Gateway\r\n\r\n"))
		clientConn.Close()
		return
	}

	klog("HTTPS: Connected to the remote server [%s]", serverConn.RemoteAddr().String())

	// Since now we have a TCP connection to the remote server,
	// return "200 OK" to the client's CONNECT method.
	clientConn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	// Transfer TCP stream from client to server, and vice versa.
	go func() {
		var wg sync.WaitGroup

		wg.Add(1)
		go proxy.transferData(clientConn, serverConn, &wg)
		wg.Add(1)
		go proxy.transferData(serverConn, clientConn, &wg)

		wg.Wait()

		clientConn.Close()
		serverConn.Close()

		klog("HTTPS: A request to the remote server [%s] was done", serverConn.RemoteAddr().String())
	}()
}

func (proxy *ProxyServer) transferData(src net.Conn, dst net.Conn, wg *sync.WaitGroup) {
	if _, err := io.Copy(dst, src); err != nil {
		klog("Failed to transfer data from %s to %s: %v",
			src.RemoteAddr().String(), dst.RemoteAddr().String(), err)
	}

	wg.Done()
}

func (proxy *ProxyServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	logRequest("HTTPS CONNECT", r)

	if proxy.mitmMode {
		proxy.handleHTTPSWithMITM(w, r)
	} else {
		proxy.handleHTTPS(w, r)
	}
}
