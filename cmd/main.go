package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kitauji/kproxy"
	kcert "github.com/kitauji/kutils/certificate"
)

const (
	DefaultProxyPort = ":8080"
)

func main() {
	proxyPort := flag.String("port", DefaultProxyPort, "Proxy server's port such \":9999\"")
	mitm := flag.Bool("mitm", false, "Enalbe/disable MITM(man-in-the-middle) mode. Default is false.")
	caCertFile := flag.String("cacert", "", "CA's cert file path")
	caKeyFile := flag.String("cakey", "", "CA's private key file path")
	flag.Parse()

	// if mitm option is enabled, load CA files
	if *mitm {
		if *caCertFile == "" || *caKeyFile == "" {
			fmt.Fprintf(os.Stderr, "Both ca-cert and ca-key opitons are necessary")
			os.Exit(1)
		}

		if err := kcert.LoadCA(*caCertFile, *caKeyFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load CA files : %v", err)
			os.Exit(1)
		}
	}

	// Create and start a proxy server on port 8080
	kproxy.EnableLogging()
	proxy := kproxy.NewProxyServer(*mitm)

	// Configure hook handlers for just testing purpose
	proxy.OnBeforeRequest = func(req *http.Request) *http.Request {
		req.Header.Set("X-KPROXY-REQUEST", "ABC")
		return req
	}

	proxy.OnBeforeResponse = func(resp *http.Response) *http.Response {
		resp.Header.Set("X-KPROXY-RESPONSE", "XYZ")
		return resp
	}

	if err := http.ListenAndServe(*proxyPort, proxy); err != nil {
		log.Printf("Error : %v", err)
		os.Exit(1)
	}
}
