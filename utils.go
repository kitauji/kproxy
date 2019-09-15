package kproxy

import (
	"net/http"
	"strings"
)

func logRequest(desc string, r *http.Request) {
	klog("----- %s -----", desc)
	klog("  Method [%s]", r.Method)
	klog("  Host [%s]", r.Host)
	klog("  URL [%s]", r.URL.String())
	klog("  Proto [%s]", r.Proto)
	klog("  RemoteAddr [%s]", r.RemoteAddr)
	klog("  RequestURI [%s]", r.RequestURI)
	for header, values := range r.Header {
		klog("  Header: [%s] : %v", header, values)
	}
}

func logResponse(desc string, resp *http.Response) {
	klog("----- %s -----", desc)
	klog("  Status [%s]", resp.Status)
	klog("  Status Code [%d]", resp.StatusCode)
	klog("  ContentLength [%d]", resp.ContentLength)
	klog("  Proto [%s]", resp.Proto)
	for header, values := range resp.Header {
		klog("  Header : [%s] : %v", header, values)
	}
	for trailer, values := range resp.Trailer {
		klog("  Trailer: [%s] : %v", trailer, values)
	}
}

func copyHeaders(dstHeader, srcHeader http.Header) {
	for key, values := range srcHeader {
		for _, value := range values {
			dstHeader.Add(key, value)
		}
	}
}

func copyHeadersForProxy(dstHeader, srcHeader http.Header) {
	copyHeaders(dstHeader, srcHeader)
	removeProxyHeaders(dstHeader)
}

func removeProxyHeaders(h http.Header) {
	h.Del("Accept-Encoding")
	h.Del("Proxy-Connection")
	h.Del("Proxy-Authenticate")
	h.Del("Proxy-Authorization")
	h.Del("Connection")
}

func httpError(w http.ResponseWriter, status string, statusCode int, err error) {
	klog("%s : %v", status, err)
	http.Error(w, status, statusCode)
}

func parseHostname(host string) string {
	if !strings.Contains(host, ":") {
		return host
	}

	hostPort := strings.Split(host, ":")
	return hostPort[0]
}
